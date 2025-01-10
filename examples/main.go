package main

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	"github.com/nxtcoder17/ivy"
	"github.com/nxtcoder17/ivy/middleware"
)

func main() {
	r := ivy.NewRouter()

	r.Get("/hi", func(c *ivy.Context) error {
		return c.SendString("hello")
	})

	mw := func(c *ivy.Context) error {
		dir := c.QueryParam("dir")
		if dir == "" {
			return fmt.Errorf("invalid query-param (dir = %q)", dir)
		}

		dir, err := filepath.Abs(dir)
		if err != nil {
			return err
		}

		c.KV.Set("dir", dir)
		return c.Next()
	}

	r.Get("/hi-qp", middleware.MustHaveQueryParams("message", "dir"), mw, func(c *ivy.Context) error {
		return c.SendString(fmt.Sprintf("dir: %s, message: %s", c.KV.Get("dir"), c.QueryParam("message")))
	})

	if err := http.ListenAndServe(":8089", r); err != nil {
		log.Fatal(err)
	}
}
