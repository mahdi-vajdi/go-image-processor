package router

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mahdi-vajdi/go-image-processor/internal/handler"
	"github.com/mahdi-vajdi/go-image-processor/internal/middleware"
)

type Router struct {
	*mux.Router
	handler handler.Handler
}

func New(handler handler.Handler) *Router {
	r := &Router{
		Router:  mux.NewRouter(),
		handler: handler,
	}

	r.Use(middleware.Logging)
	r.Use(middleware.Recovery)

	r.setupRoutes()

	return r
}

func (r *Router) setupRoutes() {
	api := r.PathPrefix("/api").Subrouter()

	apiV1 := api.PathPrefix("/v1").Subrouter()

	publicApiV1 := apiV1.PathPrefix("/public").Subrouter()
	publicApiV1.HandleFunc("/ping", r.handler.Ping).Methods(http.MethodGet)

	imageApiV1 := apiV1.PathPrefix("/image").Subrouter()
	imageApiV1.HandleFunc("/upload", r.handler.UploadImage).Methods(http.MethodPost)
	imageApiV1.HandleFunc("/status/{taskId}", r.handler.GetImageStatus).Methods(http.MethodGet)
	imageApiV1.HandleFunc("/{imageKey}", r.handler.GetImage).Methods(http.MethodGet)
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.Router.ServeHTTP(w, req)
}
