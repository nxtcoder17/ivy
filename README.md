## IVY: net/http based router implementation

### Usage

```go
router := ivy.NewRouter()

// for middlewares
router.Use(...)

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
router.Mount("/v2/", router2)

// start server with ivy route, just like mux
http.ListenAndServe(addr, router)
```
