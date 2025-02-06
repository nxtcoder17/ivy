package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/nxtcoder17/ivy"
	"github.com/nxtcoder17/ivy/middleware"
)

func main() {
	r := ivy.NewRouter(ivy.WithErrorHandler(func(err error, w http.ResponseWriter, r *http.Request) {
		http.Error(w, fmt.Sprintf("[ERROR HANDLER]: %s", err.Error()), 500)
	}))

	r.Use(middleware.Logger())

	r.Get("/_ping", func(c *ivy.Context) error {
		return c.SendString("hi")
	})

	// sub router
	r2 := ivy.NewRouter()
	r2.Get("/_ping",
		// middleware 1
		func(c *ivy.Context) error {
			fmt.Println("INSIDE middleware 1")
			c.KV.Set("hello", "middleware 1")
			return c.Next()
		},

		// middleware 2
		func(c *ivy.Context) error {
			fmt.Println("INSIDE middleware 2")
			c.KV.Set("world", "middleware 2")
			return c.Next()
		},

		// handler
		func(c *ivy.Context) error {
			<-time.After(1 * time.Second)
			return c.SendString(fmt.Sprintf("OK! from router 2 (hello = %v, world = %v)", c.KV.Get("hello"), c.KV.Get("world")))
		},
	)

	r.Mount("/v2", r2)

	r.Get("/error", func(c *ivy.Context) error {
		return fmt.Errorf("handler error")
	})

	addr := ":8089"
	slog.Info("http server started", "addr", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal(err)
	}
}
