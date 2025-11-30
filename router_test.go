package silverlining

import (
	"bytes"
	"testing"
)

func TestAppBasicRoute(t *testing.T) {
	app := New()

	var called bool
	app.Get("/hello", func(c *Context) {
		called = true
	})

	ctx := newTestContext(MethodGET, "/hello")
	defer PutRequestContext(ctx)

	app.Handler()(ctx)

	if !called {
		t.Fatal("handler was not executed")
	}
}

func TestAppRouteParams(t *testing.T) {
	app := New()

	app.Get("/users/:id/books/:book", func(c *Context) {
		if got := c.Params("id"); got != "42" {
			t.Fatalf("expected id=42, got %q", got)
		}
		if got := c.Params("book"); got != "go-internals" {
			t.Fatalf("expected book=go-internals, got %q", got)
		}
	})

	ctx := newTestContext(MethodGET, "/users/42/books/go-internals")
	defer PutRequestContext(ctx)

	app.Handler()(ctx)
}

func TestAppWildcardRoute(t *testing.T) {
	app := New()

	app.Get("/static/*file", func(c *Context) {
		if got := c.Params("file"); got != "css/app.css" {
			t.Fatalf("expected css/app.css, got %q", got)
		}
	})

	ctx := newTestContext(MethodGET, "/static/css/app.css")
	defer PutRequestContext(ctx)

	app.Handler()(ctx)
}

func TestAppMiddlewareOrder(t *testing.T) {
	app := New()

	var order []string

	app.Use(func(c *Context) {
		order = append(order, "global-before")
		c.Next()
		order = append(order, "global-after")
	})

	group := app.Group("/api", func(c *Context) {
		order = append(order, "group")
		c.Next()
	})

	group.Get("/v1/users", func(c *Context) {
		order = append(order, "handler")
	})

	ctx := newTestContext(MethodGET, "/api/v1/users")
	defer PutRequestContext(ctx)

	app.Handler()(ctx)

	expected := []string{"global-before", "group", "handler", "global-after"}
	if len(order) != len(expected) {
		t.Fatalf("unexpected order length: %v", order)
	}
	for i := range expected {
		if order[i] != expected[i] {
			t.Fatalf("unexpected order %v != %v", order, expected)
		}
	}
}

func TestAppMethodNotAllowed(t *testing.T) {
	app := New()

	app.Get("/only-get", func(c *Context) {})

	ctx := newTestContext(MethodPOST, "/only-get")
	defer PutRequestContext(ctx)

	app.Handler()(ctx)

	if ctx.response.StatusCode != 405 {
		t.Fatalf("expected 405, got %d", ctx.response.StatusCode)
	}
}

func TestAppNotFound(t *testing.T) {
	app := New()

	ctx := newTestContext(MethodGET, "/missing")
	defer PutRequestContext(ctx)

	app.Handler()(ctx)

	if ctx.response.StatusCode != 404 {
		t.Fatalf("expected 404, got %d", ctx.response.StatusCode)
	}
}

func newTestContext(method Method, target string) *Context {
	buf := &bytes.Buffer{}
	ctx := GetRequestContext(buf)
	ctx.reqR.Request.Method = method
	ctx.reqR.Request.RawURI = stringToBytes(target)
	ctx.reqR.Request.URI.Parse(stringToBytes(target))
	ctx.response.reset()
	return ctx
}
