package controllers

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"tr1sm0s1n/tda/config"
	"tr1sm0s1n/tda/models"

	"github.com/gofiber/fiber/v2"
)

func CreateOne(c *fiber.Ctx, db *config.DBConn) error {
	var sinner models.Sinner
	if err := c.BodyParser(&sinner); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	result := db.Postgres.Create(&sinner)
	if result.Error != nil {
		return c.Status(fiber.StatusBadRequest).SendString(result.Error.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(sinner)
}

func ReadAll(c *fiber.Ctx, db *config.DBConn) error {
	var sinners []models.Sinner
	result := db.Postgres.Find(&sinners)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(result.Error.Error())
	}

	return c.Status(fiber.StatusOK).JSON(sinners)
}

func ReadOne(c *fiber.Ctx, db *config.DBConn) error {
	var sinner models.Sinner
	param := c.Params("code")
	code, err := strconv.Atoi(param)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	value := db.Redis.Get(context.Background(), param).Val()
	if len(value) != 0 {
		json.Unmarshal([]byte(value), &sinner)
		return c.Status(fiber.StatusOK).JSON(sinner)
	}

	result := db.Postgres.First(&sinner, "code = ?", code)
	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).SendString("Not Found")
	}

	data, _ := json.Marshal(sinner)
	if err := db.Redis.Set(context.Background(), param, data, 0).Err(); err != nil {
		log.Printf("\033[31m[ERR]\033[0m Redis Error: %v", err)
	}

	return c.Status(fiber.StatusOK).JSON(sinner)
}

func UpdateOne(c *fiber.Ctx, db *config.DBConn) error {
	var sinner models.Sinner
	param := c.Params("code")
	code, err := strconv.Atoi(param)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	result := db.Postgres.First(&sinner, "code = ?", code)
	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).SendString("Not Found")
	}

	if err := c.BodyParser(&sinner); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	result = db.Postgres.Save(&sinner)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(result.Error.Error())
	}

	return c.Status(fiber.StatusOK).JSON(sinner)
}

func DeleteOne(c *fiber.Ctx, db *config.DBConn) error {
	var sinner models.Sinner
	param := c.Params("code")
	code, err := strconv.Atoi(param)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	result := db.Postgres.First(&sinner, "code = ?", code)
	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).SendString("Not Found")
	}

	result = db.Postgres.Delete(&models.Sinner{}, code)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(result.Error.Error())
	}

	return c.Status(fiber.StatusOK).JSON(sinner)
}
