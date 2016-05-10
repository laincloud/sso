package server

import (
	"encoding/json"
	"net/http"

	"golang.org/x/net/context"
)

// ResourceHandler is a function type for the restful resources to define a json restful api
type ResourceHandler func(ctx context.Context, r *http.Request) (code int, data interface{})

// RestfulHandlerAdapter is a function type to adapt a ResourceHandler to Handler
type RestfulHandlerAdapter func(handle ResourceHandler) Handler

// Resouce is an interface to define the basic restful api entry points
type Resource interface {
	Get(ctx context.Context, r *http.Request) (code int, data interface{})
	Post(ctx context.Context, r *http.Request) (code int, data interface{})
	Put(ctx context.Context, r *http.Request) (code int, data interface{})
	Delete(ctx context.Context, r *http.Request) (code int, data interface{})
	Patch(ctx context.Context, r *http.Request) (code int, data interface{})
	Head(ctx context.Context, r *http.Request) (code int, data interface{})
}

// RestfulHandlerAdapter will set the server's restful handler adapter
func (s *Server) RestfulHandlerAdapter(adapter RestfulHandlerAdapter) {
	if adapter != nil {
		s.restfulAdapter = adapter
	}
}

// AddRestfulResource will register the resource to the path with given routing name
func (s *Server) AddRestfulResource(path string, name string, resource Resource) {
	adapter := s.restfulAdapter
	if adapter == nil {
		adapter = s.defaultRestfulAdapter
	}
	s.Get(path, "Get_"+name, adapter(resource.Get))
	s.Post(path, "Post_"+name, adapter(resource.Post))
	s.Delete(path, "Delete_"+name, adapter(resource.Delete))
	s.Put(path, "Put_"+name, adapter(resource.Put))
	s.Patch(path, "Patch_"+name, adapter(resource.Patch))
	s.Head(path, "Head_"+name, adapter(resource.Head))
}

func (s *Server) defaultRestfulAdapter(handle ResourceHandler) Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) context.Context {
		status, v := handle(ctx, r)
		data, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			status = http.StatusInternalServerError
			data = []byte(err.Error())
		}
		data = append(data, '\n')
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(status)
		if status != http.StatusNoContent {
			w.Write(data)
		}
		return ctx
	}
}

// BaseResource is a stub Resource definition with empty implemetation
type BaseResource struct{}

func (ur BaseResource) Get(ctx context.Context, r *http.Request) (int, interface{}) {
	return http.StatusMethodNotAllowed, "Get method is not supported"
}

func (ur BaseResource) Post(ctx context.Context, r *http.Request) (int, interface{}) {
	return http.StatusMethodNotAllowed, "Post method is not supported"
}

func (ur BaseResource) Delete(ctx context.Context, r *http.Request) (int, interface{}) {
	return http.StatusMethodNotAllowed, "Delete method is not supported"
}

func (ur BaseResource) Put(ctx context.Context, r *http.Request) (int, interface{}) {
	return http.StatusMethodNotAllowed, "Put method is not supported"
}

func (ur BaseResource) Patch(ctx context.Context, r *http.Request) (int, interface{}) {
	return http.StatusMethodNotAllowed, "Patch method is not supported"
}

func (ur BaseResource) Head(ctx context.Context, r *http.Request) (int, interface{}) {
	return http.StatusMethodNotAllowed, "Head method is not supported"
}
