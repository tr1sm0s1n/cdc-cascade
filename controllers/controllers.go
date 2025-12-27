package controllers

import (
	"encoding/json"
	"log"
	"strconv"
	"tr1sm0s1n/cdc-cascade/config"
	"tr1sm0s1n/cdc-cascade/models"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Controllers struct {
	pg    *gorm.DB
	redis *redis.Client
}

func NewControllers(db *config.DBConn) *Controllers {
	return &Controllers{pg: db.Postgres, redis: db.Redis}
}

func (ct *Controllers) CreateOne(c *fiber.Ctx) error {
	var sinner models.Sinner
	if err := c.BodyParser(&sinner); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	result := ct.pg.Create(&sinner)
	if result.Error != nil {
		return c.Status(fiber.StatusBadRequest).SendString(result.Error.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(sinner)
}

func (ct *Controllers) ReadAll(c *fiber.Ctx) error {
	var sinners []models.Sinner
	result := ct.pg.Find(&sinners)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(result.Error.Error())
	}

	return c.Status(fiber.StatusOK).JSON(sinners)
}

func (ct *Controllers) ReadOne(c *fiber.Ctx) error {
	var sinner models.Sinner
	param := c.Params("code")
	code, err := strconv.Atoi(param)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	value := ct.redis.Get(c.Context(), param).Val()
	if len(value) != 0 {
		json.Unmarshal([]byte(value), &sinner)
		return c.Status(fiber.StatusOK).JSON(sinner)
	}

	result := ct.pg.First(&sinner, "code = ?", code)
	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).SendString("Not Found")
	}

	data, _ := json.Marshal(sinner)
	if err := ct.redis.Set(c.Context(), param, data, 0).Err(); err != nil {
		log.Printf("\033[31m[ERR]\033[0m Redis Error: %v", err)
	}

	return c.Status(fiber.StatusOK).JSON(sinner)
}

func (ct *Controllers) UpdateOne(c *fiber.Ctx) error {
	var sinner models.Sinner
	param := c.Params("code")
	code, err := strconv.Atoi(param)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	result := ct.pg.First(&sinner, "code = ?", code)
	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).SendString("Not Found")
	}

	if err := c.BodyParser(&sinner); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	result = ct.pg.Save(&sinner)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(result.Error.Error())
	}

	return c.Status(fiber.StatusOK).JSON(sinner)
}

func (ct *Controllers) DeleteOne(c *fiber.Ctx) error {
	var sinner models.Sinner
	param := c.Params("code")
	code, err := strconv.Atoi(param)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	result := ct.pg.First(&sinner, "code = ?", code)
	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).SendString("Not Found")
	}

	result = ct.pg.Delete(&models.Sinner{}, code)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(result.Error.Error())
	}

	return c.Status(fiber.StatusOK).JSON(sinner)
}
