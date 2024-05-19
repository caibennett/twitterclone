package main

import (
	"fmt"
	"os"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
)

func main() {
	err := run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
func run() error {
	app := fiber.New()

	app.Use(logger.New(), recover.New())

	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("hello world")
	})

	app.Listen(":3000", fiber.ListenConfig{EnablePrefork: true})
	return nil
}
