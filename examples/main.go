package main

import (
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"path/filepath"
	"time"

	"github.com/nxtcoder17/ivy"
	"github.com/nxtcoder17/ivy/middleware"
	"github.com/nxtcoder17/ivy/middleware/logger"
)

func main() {
	r := ivy.NewRouter()

	r.Use(logger.New())

	r.Get("/hi", func(c *ivy.Context) error {
		<-time.After(time.Duration(rand.IntN(3)) * time.Second)
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
