package web

// Middleware is a type definition that expects
// and returns a Handler. It allows you to define middleware functions
// that can intercept HTTP requests, perform actions before or after
// passing control to another handler, modify the request or response, 
// or short-circuit the request by not calling the next handler.
type Middleware func(Handler) Handler