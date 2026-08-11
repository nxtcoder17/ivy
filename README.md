## IVY: net/http based router implementation with fiber like API

### Features

- Standard HTTP methods based functions that allow easy creation of routes
- Allows returning error from handlers, and handling it globally, instead of doing `http.Error(w, err.Error(), 500)` everywhere in your handler
- SubRouters and seamless mounting into one another
- Simpler Abstractions to write HTTP responses
- Middleware support
- Request Level Key-Value store to pass data from a middleware to next middleware

### Usage

```go
router := ivy.NewRouter()

// optional, if want to change default error handler
router.ErrorHandler = func(c *ivy.Context, err error) {
    c.Status(500).SendString(fmt.Sprintf("[ERROR HANDLER]: %s", err.Error()))
}

// for middlewares
router.Use(func (c *ivy.Context) error {
	return c.Next()
})

// http methods
router.Get("/ping", func(c *ivy.Context) error {
	return c.SendString("OK ! from router 1")
})

// router mounting
router2 := ivy.NewRouter()
router2.Get("/_ping",
	// middleware 1
	func(c *ivy.Context) error {
		logger.Info("INSIDE middleware 1")
		c.KV.Set("hello", "world")
		return c.Next()
	},

	// middleware 2
	func(c *ivy.Context) error {
		logger.Info("INSIDE middleware 2")
		return c.Next()
	},

	// handler
	func(c *ivy.Context) error {
		return c.SendString(fmt.Sprintf("OK! from router 2 (hello = %v)", c.KV.Get("hello")))
	},
)

// mouting router into another router
router.Mount("/v2", router2)

// start server with ivy route, just like mux
http.ListenAndServe(":8080", router)
```

### Middlewares

Middlewares live in their own sub-packages under `middleware/`, each exposing a `New()` constructor (and occasionally typed accessors, like `sse.Stream`).

```go
import (
    "github.com/nxtcoder17/ivy"
    "github.com/nxtcoder17/ivy/middleware/basicauth"
    "github.com/nxtcoder17/ivy/middleware/logger"
    "github.com/nxtcoder17/ivy/middleware/queryparams"
    "github.com/nxtcoder17/ivy/middleware/requestid"
    "github.com/nxtcoder17/ivy/middleware/sse"
)

router := ivy.NewRouter()

router.Use(requestid.New())
router.Use(logger.New())

// basic auth on the whole router
router.Use(basicauth.New("Restricted", map[string]string{
    "admin": "secret",
}))

// query param validation, scoped to a single route
router.Get("/search", queryparams.Required("q"), func(c *ivy.Context) error {
    return c.SendString("searching...")
})

// server-sent events
router.Get("/events", sse.New(), func(c *ivy.Context) error {
    stream, err := sse.Stream(c)
    if err != nil {
        return err
    }
    return stream.SendEvent("message", []byte("hello"))
})
```

### Examples

- [Sub Router](./examples/sub-router)
