package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
	"tr1sm0s1n/cdc-cascade/config"
	"tr1sm0s1n/cdc-cascade/controllers"
	"tr1sm0s1n/cdc-cascade/queue"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	var wg sync.WaitGroup
	c := make(chan os.Signal, 1)

	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	ctx, shutdown := context.WithCancel(context.Background())

	db, err := config.NewDBConn(ctx)
	if err != nil {
		log.Fatalf("\033[31m[ERR]\033[0m Error: %v", err)
	}

	cdc := &queue.CDC{
		Ctx:   ctx,
		Wg:    &wg,
		Redis: db.Redis,
	}

	wg.Add(1)
	go queue.Runner(cdc, queue.StartCDC)

	app := setupApp(db)

	go func() {
		if err := app.Listen(":" + os.Getenv("API_PORT")); err != nil {
			log.Printf("\033[31m[ERR]\033[0m Server crashed: %v", err)
		}
	}()

	<-c
	shutdown()

	log.Println("\033[33m[WRN]\033[0m Shutting down server...")
	if err := app.ShutdownWithTimeout(5 * time.Second); err != nil {
		log.Printf("\033[31m[ERR]\033[0m Server shutdown error: %v", err)
	}

	log.Println("\033[33m[WRN]\033[0m Waiting for consumers...")
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("\033[32m[INF]\033[0m All consumers stopped gracefully.")
	case <-time.After(5 * time.Second):
		log.Println("\033[31m[ERR]\033[0m Timeout waiting for consumers to stop.")
	}

	log.Println("\033[32m[INF]\033[0m Server shutdown complete.")
}

func setupApp(db *config.DBConn) *fiber.App {
	app := fiber.New(fiber.Config{
		ReadTimeout: 5 * time.Second,
	})
	app.Use(logger.New())

	ct := controllers.NewControllers(db)

	api := app.Group("/api")
	{
		v1 := api.Group("/v1")
		{
			sinnerRoutes := v1.Group("/sinners")
			{
				sinnerRoutes.Post("/create", ct.CreateOne)
				sinnerRoutes.Get("/read", ct.ReadAll)
				sinnerRoutes.Get("/read/:code", ct.ReadOne)
				sinnerRoutes.Put("/update/:code", ct.UpdateOne)
				sinnerRoutes.Delete("/delete/:code", ct.DeleteOne)
			}
		}
	}

	return app
}
