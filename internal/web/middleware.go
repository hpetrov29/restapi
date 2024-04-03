package web

// Middleware is a type definition that expects
// and returns a Handler. It allows you to define middleware functions
// that can intercept HTTP requests, perform actions before or after
// passing control to another handler, modify the request or response,
// or short-circuit the request by not calling the next handler.
type Middleware func(Handler) Handler

// wrapMiddleware creates a new handler by wrapping middleware around a final
// handler. The middlewares' Handlers will be executed by requests in the order
// they are provided.
func wrapMiddleware(mw []Middleware, handler Handler) Handler {

	// Loop backwards through the middleware invoking each one. Replace the
	// handler with the new wrapped handler. Looping backwards ensures that the
	// first middleware of the slice is the first to be executed by requests.
	for i := len(mw) - 1; i >= 0; i-- {
		mwFunc := mw[i]
		if mwFunc != nil {
			handler = mwFunc(handler)
		}
	}

	return handler
}
