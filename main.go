package main

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
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

	app := fiber.New(fiber.Config{})

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
		if err != nil {
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
	app.Get("/search", func(c fiber.Ctx) error {
		session := c.Cookies("session")
		if session == "" {
			return Render(c, templ.Search())
		}

		user, err := executor.GetUserFromSession(ctx, session)
		if err != nil {
			return Render(c, templ.Search())
		}
		return Render(c, templ.Search(user))
	})
	app.Get("/dosearch", func(c fiber.Ctx) error {
		_type := c.Query("t")
		trimmedQuery := strings.TrimSpace(c.Query("q"))
		if _type == "people" {
			if trimmedQuery == "" {
				return Render(c, templ.People([]db.SearchPeopleRow{}))
			}
			people, err := executor.SearchPeople(ctx, db.SearchPeopleParams{Concat: c.Query("q"), Concat_2: c.Query("q")})
			if err != nil {
				fmt.Println(err.Error())
				return c.SendStatus(400)
			}
			return Render(c, templ.People(people))
		} else if _type == "posts" {
			if trimmedQuery == "" {
				return Render(c, templ.Posts([]db.SearchPostsRow{}))
			}
			posts, err := executor.SearchPosts(ctx, c.Query("q"))
			if err != nil {
				fmt.Println(err.Error())
				return c.SendStatus(400)
			}
			return Render(c, templ.Posts(posts))
		}
		return c.SendStatus(400)
	})
	app.Get("/entersearch", func(c fiber.Ctx) error {
		c.Set("HX-Redirect", "/search?query="+c.Query("q"))
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
	app.Post("/post", func(c fiber.Ctx) error {
		session := c.Cookies("session")
		if session == "" {
			return Render(c, templ.Search())
		}

		user, err := executor.GetUserFromSession(ctx, session)
		if err != nil {
			return c.SendStatus(400)
		}
		id := idGenerator.Generate()

		err = executor.CreatePost(ctx, db.CreatePostParams{ID: id.Int64(), UserID: user.ID, Content: c.FormValue("content")})
		if err != nil {
			return c.SendStatus(500)
		}
		if c.FormValue("firstPost") == "true" {
			return Render(c, templ.PostPanel([]db.ListPostsAndUsersRow{{Name: sql.NullString{Valid: true, String: "Fortnite"}, Username: "fortnite", Content: c.FormValue("content")}}))
		}
		return Render(c, templ.Post(db.ListPostsAndUsersRow{ID: id.Int64(), UserID: user.ID, Content: c.FormValue("content"), CreatedAt: time.Now().Unix(), Name: user.Name, Username: user.Username}))
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
