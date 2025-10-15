package queue

import (
	"context"
	"errors"
	"log"
	"runtime/debug"
	"time"
)

func Runner[T any](t T, f func(T) error) {
	for {
		err := func() (err error) {
			defer func() {
				if v := recover(); v != nil {
					log.Printf("\033[31m[ERR]\033[0m Panic occured: %v\nStack: %s\n", v, debug.Stack())
					time.Sleep(2 * time.Second)
					err = errors.New("recovered from panic")
				}
			}()
			return f(t)
		}()

		if err != nil {
			switch {
			case errors.Is(err, context.Canceled):
				log.Println("\033[32m[INF]\033[0m Context cancelled. Stopping gracefully...")
				return
			default:
				log.Printf("\033[31m[ERR]\033[0m Runner crashed: %s. Restarting after 5 seconds...\n", err)
				time.Sleep(5 * time.Second)
			}
		}
	}
}
