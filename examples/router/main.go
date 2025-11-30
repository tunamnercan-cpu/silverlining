package main

import (
	"log"

	"github.com/go-www/silverlining"
)

func main() {
	app := silverlining.New()

	app.Use(func(c *silverlining.Context) {
		c.ResponseHeaders().Set("X-Powered-By", "silverlining")
		c.Next()
	})

	api := app.Group("/api", func(c *silverlining.Context) {
		c.ResponseHeaders().Set("Content-Type", "application/json")
		c.Next()
	})

	api.Get("/hello/:name", func(c *silverlining.Context) {
		name := c.Params("name", "world")
		_ = c.WriteFullBodyString(200, `{"message":"Hello, `+name+`"}`)
	})

	log.Fatal(app.Listen(":8080"))
}
