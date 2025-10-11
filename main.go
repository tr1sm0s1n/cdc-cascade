package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	"tr1sm0s1n/tda/config"
	"tr1sm0s1n/tda/controllers"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	c := make(chan os.Signal, 1)

	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	ctx, shutdown := context.WithCancel(context.Background())

	db, err := config.NewDBConn(ctx)
	if err != nil {
		log.Fatalf("\033[31m[ERR]\033[0m Error: %v", err)
	}

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

	log.Println("\033[33m[INF]\033[0m Server shutdown complete.")
}

func setupApp(db *config.DBConn) *fiber.App {
	app := fiber.New(fiber.Config{
		ReadTimeout: 5 * time.Second,
	})
	app.Use(logger.New())

	api := app.Group("/api")
	{
		v1 := api.Group("/v1")
		{
			sinnerRoutes := v1.Group("/sinners")
			{
				sinnerRoutes.Post("/create", func(ctx *fiber.Ctx) error {
					return controllers.CreateOne(ctx, db)
				})
				sinnerRoutes.Get("/read", func(ctx *fiber.Ctx) error {
					return controllers.ReadAll(ctx, db)
				})
				sinnerRoutes.Get("/read/:code", func(ctx *fiber.Ctx) error {
					return controllers.ReadOne(ctx, db)
				})
				sinnerRoutes.Put("/update/:code", func(ctx *fiber.Ctx) error {
					return controllers.UpdateOne(ctx, db)
				})
				sinnerRoutes.Delete("/delete/:code", func(ctx *fiber.Ctx) error {
					return controllers.DeleteOne(ctx, db)
				})
			}
		}
	}

	return app
}
