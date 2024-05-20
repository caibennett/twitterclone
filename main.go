package main

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"twitterclone/db"
	"twitterclone/templ"

	"github.com/bwmarrin/snowflake"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	err := run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
func run() error {
	ctx := context.Background()
	snowflake.Epoch = 1712569420
	idGenerator, err := snowflake.NewNode(1)
	if err != nil {
		return err
	}

	executor, err := connectToDb()
	if err != nil {
		return err
	}

	app := fiber.New()

	app.Use(logger.New(), recover.New())
	app.Static("/assets", "./assets")

	app.Get("/", func(c fiber.Ctx) error {
		posts, err := executor.ListPostsAndUsers(ctx)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return c.SendString("Failed to load posts")
		}
		session := c.Cookies("session")
		if session == "" {
			return Render(c, templ.Main(posts))
		}

		user, err := executor.GetUserFromSession(ctx, session)
		if errors.Is(err, sql.ErrNoRows) {
			fmt.Println("Session not found")
			return Render(c, templ.Main(posts))
		}
		return Render(c, templ.Main(posts, user))
	})
	app.Get("/onboarding", func(c fiber.Ctx) error {
		session := c.Cookies("session")
		if session == "" {
			return c.Redirect().To("/")
		}
		_, err := executor.GetUserFromSession(ctx, session)
		if errors.Is(err, sql.ErrNoRows) {
			return c.Redirect().To("/")
		}
		return Render(c, templ.Onboarding())
	})
	app.Put("/set_name", func(c fiber.Ctx) error {
		session := c.Cookies("session")
		if session == "" {
			return c.SendString("Unauthorised, please refresh")
		}
		user, err := executor.GetUserFromSession(ctx, session)
		if errors.Is(err, sql.ErrNoRows) {
			return c.SendString("Unauthorised, please refresh")
		}
		name := c.FormValue("displayname")
		if len(name) < 3 {
			return c.SendString("Name must be longer than 3 letters")
		}

		err = executor.SetName(ctx, db.SetNameParams{Name: sql.NullString{String: name, Valid: true}, ID: user.ID})
		if err != nil {
			return c.SendString("Something went wrong")
		}
		c.Set("HX-Redirect", "/")
		return c.SendStatus(200)
	})
	app.Get("/sign/out", func(c fiber.Ctx) error {
		session := c.Cookies("session")
		if session == "" {
			return c.Redirect().To("/")
		}
		executor.DeleteSession(ctx, session)
		ClearSession(c)
		return c.Redirect().To("/")
	})
	app.Get("/sign/in", func(c fiber.Ctx) error {
		return Render(c, templ.SignIn())
	})

	app.Post("/sign/in", func(c fiber.Ctx) error {
		session := c.Cookies("session")
		if session != "" {
			executor.DeleteSession(ctx, session)
		}

		username := c.FormValue("username")
		password := c.FormValue("password")

		if len(username) < 3 || len(password) < 8 {
			return c.SendString("Does not meet the minimum size")
		}

		user, err := executor.GetFromUsername(ctx, username)
		if errors.Is(err, sql.ErrNoRows) {
			return c.SendString("Incorrect username")
		}
		if err != nil {
			fmt.Println(err.Error())
			return c.SendString("Something went wrong")
		}

		valid := CompareHash(password, user.Password)
		if !valid {
			return c.SendString("Incorrect Password")
		}

		token, err := CreateSession(executor, c.IP(), user.ID, true)
		if err != nil {
			fmt.Println(err.Error())
			return c.SendString("Something went wrong")
		}
		SetSession(c, token)
		c.Set("HX-Redirect", "/")
		return c.SendStatus(200)
	})

	app.Get("/sign/up", func(c fiber.Ctx) error {
		return Render(c, templ.SignUp())
	})
	app.Post("/sign/up", func(c fiber.Ctx) error {
		username := c.FormValue("username")
		password := c.FormValue("password")

		if len(username) < 3 || len(password) < 8 {
			return c.SendString("Does not meet the minimum size")
		}
		_, err := executor.GetUsername(ctx, username)

		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return c.SendString("Username is taken")
		}

		id := idGenerator.Generate()
		hash, err := HashPassword(password)
		if err != nil {
			fmt.Println(err.Error())
			return c.SendString("Something went wrong")
		}
		err = executor.CreateUser(ctx, db.CreateUserParams{ID: id.Int64(), Username: username, Password: hash})
		if err != nil {
			fmt.Println(err.Error())
			return c.SendString("Something went wrong")
		}
		token, err := CreateSession(executor, c.IP(), id.Int64(), false)
		if err != nil {
			fmt.Println(err.Error())
			return c.SendString("Something went wrong")
		}
		SetSession(c, token)
		c.Set("HX-Redirect", "/onboarding")
		return c.SendStatus(200)
	})
	app.Listen(":3000", fiber.ListenConfig{EnablePrefork: true})
	return nil
}
func connectToDb() (*db.Queries, error) {
	sql, err := sql.Open("sqlite3", "app.db")
	if err != nil {
		return nil, err
	}
	return db.New(sql), nil
}
