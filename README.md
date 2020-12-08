# Go Router

```go
type Router struct {
    NotFoundHandler http.Handler
}
```

## Setting path params

Register routes with Router.AddRoute. Parts of the URI which start with a ":" are parameters. For example, if you registered `/user/:userID`, it would match `example.com/user/3`. You would then get a param called `userID` which would equal `"3"`.

Inside the handler function you provide, access path parameters with the `PathParams` function.

```go
router.AddRoute(http.MethodGet, "user/:userID", func(w http.ResponseWriter, r *http.Request) {
    params, _ := router.PathParams(requestMethod, requestPath)
    
    userID := params["userID"]
})
```

## Not Found Handler

If Router.NotFoundHandler is not a set, a default handler will be called when a route is not found. If you want to set your own handler, set Router.NotFoundHandler with the http.Handler you would prefer.
