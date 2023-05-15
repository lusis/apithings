# FAQ
I generally have questions when evaluating a new application/codebase and I figured I'd at least attempt to answer the ones I would ask

## Why is the code broken out the way it is? (provider, storer, handler)
So I'm not a traditional software developer in that I have no formal education and everything is self-taught. Additionally, my background is working in the operations space not primarily as a software developer (though I've always written tooling and software).

The point of that statement is that I tend to walk backwards into some design pattern unintentionally.

The layout of the code is the one that makes the most sense to me and allows the most flexible testing capabilities. I need good test coverage so I don't shoot myself in the foot. One reason I love writing Go is that it's generally hard to shoot yourself in the foot and the edge cases for that are fairly well known at this point.

So yes, I could bypass the service layer and just go http.HandlerFunc -> storer but that intermingles too much for me. Having the intermediary of the service (provider in this repo - service is an overloaded term) allows the handler to focus on api semantics.

Also I tend to break out types into their own packages to avoid complicated dependency loops down the road. In professional projects, I prefer to define all my types via proto3 and generate the code but that was, again, overkill for this project.

## Why not technology `X`?
In general, I try to use stdlib code as much as possible. I've found this is the most approachable for new developers and contributors to go codebases.

In cases where a third-party library is used, I like to pick ones that let you work with stdlib types as much as possible. Let's take for instance two popular http routers:

Using the basic examples from each repo's README, we'll compare [`chi`](https://github.com/go-chi/chi) and [`gin`](https://github.com/gin-gonic/gin)

_previously I would have used `gorilla.Mux` in my examples but the project has long since been archived and is unmaintained and `chi` is a spiritual successor to `gorilla.Mux` for me_

- `gin`
```go
r := gin.Default()
r.GET("/ping", func(c *gin.Context) {
c.JSON(http.StatusOK, gin.H{
    "message": "pong",
})
})
r.Run()
```

- `chi`
_note I've trimmed the optional middleware bits to make the example's line up better_
```go
r := chi.NewRouter()

r.Get("/", func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("welcome"))
})
http.ListenAndServe(":3000", r)
```

For the most part, all http routers provide http verb named functions to set allowed methods on a given request and the params are generally a path string and a function to call when that method + path combination is hit.
What's most different between these two examples, however, is what the signature of the function that will be called is:

`gin` uses `func(*gin.Context) {}` while `chi` uses `func(http.ResponseWriter, *http.Request) {}`

In the case of `chi`, I'm working with stdlib types here (`http.HandlerFunc`) and can use the go stdlib documentation to work through examples and issues.
With `gin` I have to need to understand `*gin.Context` and what's different between that and a `context.Context`.
Also note that `gin.Context` is a struct while `context.Context` is an interface which makes testing much easier.

Additionally (and this is, again a minor thing), we have to use a special function on the router to actually serve http traffic whereas with `chi`, the router we get back implements the stdlib [http.Handler](https://pkg.go.dev/net/http#Handler) so we can go right back to using it with the examples and documentation we can find online for writing http servers in go.

### sidebar on `gin`
Again I want to stress that I am not criticizing the `gin` project here or the quality of its code.
A `gin.Context` implements the `context.Context` interface and `gin` WAS around before the standardization of `context.Context` in (I think) `1.17`. Gin is also a MORE than just an http router. It's more akin to a Django in that regard.
My point here is about approachability for people new to a codebase - the same concerns are valid for any third-party library you add to your projects.


### Why did you not use `chi` initially?
So why didn't I juse use `chi` initially in this project? That's a bigger answer but it boiled down to challenging my assumptions on a new project. I could have used `chi` and it wouldn't have added much complication or overhead to the project but I didn't NEED it at the time. 

As I said above, most http routers for Go implement `http.Handler` which has ONE function: `ServeHTTP(http.ResponseWriter, *http.Request)`.
Something like `foo.Get("/", myfunc)` deep down perform the following checks:

- is [`Request.Method`](https://pkg.go.dev/net/http#Request) an `http.MethodGet`?
- is [`Request.Path`](https://pkg.go.dev/net/http#Request) `/`?

You don't need a router to do this. You don't even need a router for middleware. You can wrap your function with something else that returns an `http.HandlerFunc` like so:

```go
package main

import (
	"net/http"

	"golang.org/x/exp/slog"
)

func logWrapper(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.Info("got a request")
		next(w, r)
	}
}

func mycode(w http.ResponseWriter, r *http.Request) {
	// do a thing
	slog.Info("my request")
}

func main() {
	http.HandleFunc("/foo", logWrapper(mycode))
	http.ListenAndServe(":8080", nil) // nolint: errcheck
}
```

which results in the following when `/foo` is called:
```
2023/05/03 10:32:18 INFO got a request
2023/05/03 10:32:18 INFO my request
```

So back to challenging assumptions. I took this project as an opportunity to remind myself WHY I use routers like `chi` and `gin` by manually implementing the router functionality. I feel like this makes you a better mentor and teacher (ask me about pointer recievers sometime).

Having said that, maintaining my own router became cumbersome and made the code much less readable so I did end up migrating to `chi` internally.

### `mysql`/`postgres` support
Coming soon. Sqlite was the easiest local option and needed for one of the deploy models I'm working on.

### Much `internal`. Such sad
I mentioned this in the main README but right now I'm not ready for contributions/PRs just yet. I'd like to keep the exported surface area minimal for now while I decide what makes SENSE to export.