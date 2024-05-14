package api

import "net/http"

type IModule interface {
	GetRoutes() []Route
}

type Route struct {
	Path        string
	HandlerFunc func(w http.ResponseWriter, r *http.Request)
	Method      string
}

func NewRoute(path string, method string, handlerFunc func(w http.ResponseWriter, r *http.Request)) *Route {
	return &Route{
		Path:        path,
		Method:      method,
		HandlerFunc: handlerFunc,
	}
}
