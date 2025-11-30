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

## Benchmarks

`benchmarks/silverlining_server.go` contains the minimal plaintext handler that was used to compare this fork against [fasthttp](https://github.com/valyala/fasthttp). Each server listened on `:18080` and returned the same `Hello, World!` response. The fasthttp side used the equivalent program below:

```go
package main

import (
	"log"

	"github.com/valyala/fasthttp"
)

func main() {
	handler := func(ctx *fasthttp.RequestCtx) {
		ctx.Response.Header.SetContentType("text/plain")
		ctx.SetStatusCode(200)
		ctx.SetBodyString("Hello, World!")
	}

	log.Fatal(fasthttp.ListenAndServe(":18080", handler))
}
```

### Command

```
wrk -t50 -c50 -d30s http://127.0.0.1:18080/
```

### Environment

- Ubuntu 24.04 (kernel 6.1.147) inside KVM with 4 vCPU Intel(R) Xeon(R) Processor
- Go 1.22.2, wrk 4.1.0

### Results (`-t50 -c50`, 30s run)

| Framework | Requests/sec | Transfer/sec | Total requests |
| --- | --- | --- | --- |
| silverlining (this fork) | 392,297.87 | 37.79 MB/s | 11,792,295 |
| fasthttp v1.51.0 | 391,340.39 | 37.69 MB/s | 11,777,243 |

The two implementations are effectively neck and neck (±0.25%). Re-run the commands above on your hardware if you need numbers that reflect a different environment.
