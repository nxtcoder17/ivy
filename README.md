## IVY: net/http based router implementation with API like express.js or go-fiber

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
		return c.Next()
	},

	// middleware 2
	func(c *ivy.Context) error {
		logger.Info("INSIDE middleware 2")
		return c.Next()
	},

	// handler
	func(c *ivy.Context) error {
		return c.SendString("OK! from router 2")
	},
)

// mouting router into another router
router.Mount("/v2", router2)

// start server with ivy route, just like mux
http.ListenAndServe(addr, router)
```
