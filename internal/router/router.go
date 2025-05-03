package router

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mahdi-vajdi/go-image-processor/internal/handler"
	"github.com/mahdi-vajdi/go-image-processor/internal/middleware"
)

type Router struct {
	*mux.Router
	publicHandler handler.PublicHandler
}

func New(publicHandler handler.PublicHandler) *Router {
	r := &Router{
		Router:        mux.NewRouter(),
		publicHandler: publicHandler,
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
	publicApiV1.HandleFunc("/ping", r.publicHandler.Ping).Methods(http.MethodGet)
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.Router.ServeHTTP(w, req)
}
