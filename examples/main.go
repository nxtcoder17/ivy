package main

import (
	"log"
	"net/http"

	"github.com/nxtcoder17/ivy"
	"github.com/nxtcoder17/ivy/middleware"
)

func main() {
	r := ivy.NewRouter()

	r.Get("/hi", func(c *ivy.Context) error {
		return c.SendString("hello")
	})

	r.Get("/hi-qp", middleware.MustHaveQueryParams("message"), func(c *ivy.Context) error {
		return c.SendString("hello")
	})

	if err := http.ListenAndServe(":8089", r); err != nil {
		log.Fatal(err)
	}
}
