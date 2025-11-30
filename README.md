[![GitHub last commit](https://img.shields.io/github/last-commit/go-www/silverlining?style=for-the-badge)](https://github.com/go-www/silverlining/commits/main)
[![Go Reference](https://img.shields.io/badge/Go-Reference-007d9c?style=for-the-badge&logo=go)](https://pkg.go.dev/github.com/go-www/silverlining)

# silverlining

Silverlining is a low-level HTTP Framework for Go Programming Language.

## Installation

```sh
go get -u github.com/go-www/silverlining
```

## Usage

```go
package main

import "github.com/go-www/silverlining"

func main() {
	silverlining.ListenAndServe(":8080", func(r *silverlining.Context) {
		r.WriteFullBodyString(200, "Hello, World!")
	})
}

```

## Router Example

```go
package main

import (
	"log"

	"github.com/go-www/silverlining"
)

func main() {
	app := silverlining.New()

	app.Use(func(c *silverlining.Context) {
		// Runs before every route
		c.Next()
	})

	api := app.Group("/api", func(c *silverlining.Context) {
		// Runs before every /api route
		c.Next()
	})

	api.Get("/users/:id", func(c *silverlining.Context) {
		id := c.Params("id", "0")
		c.WriteFullBodyString(200, "user "+id)
	})

	log.Fatal(app.Listen(":8080"))
}
```

```go
// TLS example with custom ClientHello configuration.
package main

import (
	"crypto/tls"
	"log"

	"github.com/go-www/silverlining"
)

func main() {
	legacyCert, err := tls.LoadX509KeyPair("legacy.crt", "legacy.key")
	if err != nil {
		log.Fatal(err)
	}

	opts := &silverlining.TLSOptions{
		CertFile: "server.crt",
		KeyFile:  "server.key",
		CustomConfig: func(hello *tls.ClientHelloInfo) (*tls.Config, error) {
			if hello.ServerName == "legacy.local" {
				return &tls.Config{Certificates: []tls.Certificate{legacyCert}}, nil
			}
			return nil, nil
		},
	}

	log.Fatal(silverlining.ListenAndServeTLS(":8443", func(r *silverlining.Context) {
		r.WriteFullBodyString(200, "Hello, TLS!")
	}, opts))
}
```
