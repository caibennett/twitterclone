package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"
	"twitterclone/db"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"golang.org/x/crypto/bcrypt"
)

func Render(c fiber.Ctx, component templ.Component, options ...func(*templ.ComponentHandler)) error {
	componentHandler := templ.Handler(component)
	for _, o := range options {
		o(componentHandler)
	}
	return adaptor.HTTPHandler(componentHandler)(c)
}
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}
func CompareHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
func CreateSession(executor *db.Queries, ip string, userId int64, rememberMe bool) (string, error) {
	token, err := RandString(32)
	if err != nil {
		return "", err
	}
	return token, executor.CreateSession(context.Background(), db.CreateSessionParams{Token: token, UserID: userId, IpAddress: ip, ExpireAt: createExpiration(!rememberMe)})
}
func RandString(length int) (string, error) {
	randomBytes := make([]byte, length/2) // Divide by 2 because 2 hex characters represent each byte
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	randomString := hex.EncodeToString(randomBytes)
	return randomString, nil
}
func createExpiration(expire bool) int64 {
	var futureTime time.Time
	if expire {
		// 2 Weeks
		futureTime = time.Now().Add(14 * 24 * time.Hour)
	} else {
		// 6 Months
		futureTime = time.Now().Add(6 * 30 * 24 * time.Hour)
	}
	return futureTime.Unix()
}
func SetSession(c fiber.Ctx, value string) {
	c.Cookie(&fiber.Cookie{Name: "session", Path: "/", SameSite: "strict", HTTPOnly: true, Value: value, MaxAge: 34473600})
}
func ClearSession(c fiber.Ctx) {
	c.Cookie(&fiber.Cookie{Name: "session", Path: "/", SameSite: "strict", HTTPOnly: true, Value: "", MaxAge: -1})
}
