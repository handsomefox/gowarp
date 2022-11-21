package middleware

import "net/http"

// Decorator is an interface that allows wrapping http.Handler(s) with other http.Handler(s).
type Decorator interface {
	Decorate(http.Handler) http.Handler
}

type DecoratorFunc func(http.Handler) http.Handler

// wrap the given http.Handler with all the provided decorators.
func Decorate(who http.Handler, with ...DecoratorFunc) http.Handler {
	for _, w := range with {
		who = w(who)
	}
	return who
}
