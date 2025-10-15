package queue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/twmb/franz-go/pkg/kgo"
)

type CDC struct {
	Ctx   context.Context
	Wg    *sync.WaitGroup
	Redis *redis.Client
}

type Message struct {
	Payload Payload `json:"payload"`
}

type Payload struct {
	Before      *Sinner `json:"before"`
	After       *Sinner `json:"after"`
	Source      Source  `json:"source"`
	Transaction any     `json:"transaction"`
	Op          string  `json:"op"`
	TsMs        int64   `json:"ts_ms"`
	TsUs        int64   `json:"ts_us"`
	TsNs        int64   `json:"ts_ns"`
}

type Source struct {
	Version   string `json:"version"`
	Connector string `json:"connector"`
	Name      string `json:"name"`
	TsMs      int64  `json:"ts_ms"`
	Snapshot  string `json:"snapshot"`
	DB        string `json:"db"`
	Sequence  any    `json:"sequence"`
	TsUs      int64  `json:"ts_us"`
	TsNs      int64  `json:"ts_ns"`
	Schema    string `json:"schema"`
	Table     string `json:"table"`
	TxID      int64  `json:"txId"`
	Lsn       int64  `json:"lsn"`
	Xmin      any    `json:"xmin"`
}

type Sinner struct {
	Code int `json:"code"`
}

func StartCDC(cc *CDC) error {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(fmt.Sprintf("%s:%s", os.Getenv("KAFKA_HOST"), os.Getenv("KAFKA_BROKER_PORT"))),
		kgo.ConsumerGroup(os.Getenv("KAFKA_CONSUMER_GROUP")),
		kgo.ConsumeTopics(os.Getenv("KAFKA_CDC_TOPIC")),
		kgo.ConsumeStartOffset(kgo.NewOffset().AtStart()),
		kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()),
		kgo.AutoCommitMarks(),
		kgo.BlockRebalanceOnPoll(),
		kgo.RequireStableFetchOffsets(),
		kgo.SessionTimeout(30*time.Second),
		kgo.HeartbeatInterval(10*time.Second),
		kgo.RequestTimeoutOverhead(60*time.Second),
		kgo.RebalanceTimeout(60*time.Second),
		kgo.RetryBackoffFn(func(tries int) time.Duration {
			baseDelay := 3 * time.Second
			maxDelay := 60 * time.Second
			delay := min(time.Duration(1<<uint(tries-1))*baseDelay, maxDelay)
			jitterRange := int64(delay / 4)
			if jitterRange <= 0 {
				return delay
			}
			jitter := time.Duration(rand.Int63n(2*jitterRange)) - time.Duration(jitterRange)
			return delay + jitter
		}),
		kgo.FetchMaxBytes(10*1024*1024),
		kgo.FetchMinBytes(1*1024),
		kgo.FetchMaxWait(5*time.Second),
	)
	if err != nil {
		return fmt.Errorf("failed to create Kafka client: %v", err)
	}

	defer func() {
		client.AllowRebalance()

		leaveCtx, leaveCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer leaveCancel()

		if err := client.LeaveGroupContext(leaveCtx); err != nil {
			log.Printf("\033[31m[ERR]\033[0m Failed to leave consumer group: %v\n", err)
		}

		client.Close()
	}()

	log.Printf("\033[32m[INF]\033[0m Cache Clearer is running.\n")

	for {
		select {
		case <-cc.Ctx.Done():
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			if err := client.CommitMarkedOffsets(ctx); err != nil {
				log.Printf("\033[31m[ERR]\033[0m Failed to commit offsets before shutdown: %v\n", err)
			} else {
				log.Println("\033[32m[INF]\033[0m Successfully committed offsets before shutdown.")
			}
			cancel()

			cc.Wg.Done()
			return cc.Ctx.Err()

		default:
			fetches := client.PollRecords(cc.Ctx, 100)

			if fetches.IsClientClosed() {
				client.AllowRebalance()
				if errors.Is(fetches.Err0(), context.Canceled) {
					cc.Wg.Done()
				}
				return fetches.Err0()
			}

			if fetches.Empty() {
				client.AllowRebalance()
				continue
			}

			// Handle any errors
			if errs := fetches.Errors(); len(errs) > 0 {
				for _, err := range errs {
					if errors.Is(err.Err, context.Canceled) {
						client.AllowRebalance()
						cc.Wg.Done()
						return err.Err
					}
					log.Printf("\033[31m[ERR]\033[0m Fetch error: %v.\n", err)
				}
				client.AllowRebalance()
				return fmt.Errorf("too many fetch errors")
			}

			var processingErr error
			fetches.EachPartition(func(p kgo.FetchTopicPartition) {
				if processingErr != nil {
					return
				}

				for _, record := range p.Records {
					select {
					case <-cc.Ctx.Done():
						processingErr = cc.Ctx.Err()
						return
					default:
					}

					log.Printf("\033[32m[INF]\033[0m Message arrived: partition=%d, offset=%d\n", record.Partition, record.Offset)

					var message Message
					if record.Value == nil {
						client.MarkCommitRecords(record)
						continue
					}

					if err := json.Unmarshal(record.Value, &message); err != nil {
						log.Printf("\033[31m[ERR]\033[0m Failed to unmarshal record: %v\n", err)
						client.MarkCommitRecords(record) // Mark as processed despite error
						continue
					}

					if message.Payload.Before == nil {
						log.Printf("\033[32m[INF]\033[0m First entry: code=%d, partition=%d, offset=%d\n",
							message.Payload.After.Code, record.Partition, record.Offset)
						client.MarkCommitRecords(record)
						continue
					}

					if err := cc.Redis.Del(cc.Ctx, strconv.Itoa(message.Payload.Before.Code)).Err(); err != nil {
						processingErr = err
						return
					}

					log.Printf("\033[32m[INF]\033[0m Cache cleared: code=%d, partition=%d, offset=%d\n",
						message.Payload.Before.Code, record.Partition, record.Offset)

					client.MarkCommitRecords(record)
				}
			})

			// Handle processing errors
			if processingErr != nil {
				log.Printf("\033[31m[ERR]\033[0m Processing error: %v\n", processingErr)
				client.AllowRebalance()

				if errors.Is(processingErr, context.Canceled) {
					cc.Wg.Done()
				}
				return processingErr
			}

			// Commit offsets
			commitCtx, commitCancel := context.WithTimeout(context.Background(), 5*time.Second)
			err := client.CommitMarkedOffsets(commitCtx)
			commitCancel()

			if err != nil {
				log.Printf("\033[31m[ERR]\033[0m Failed to commit offsets: %v\n", err)
				client.AllowRebalance()
				return fmt.Errorf("failed to commit offsets: %v", err)
			}

			client.AllowRebalance()
		}
	}
}
