package handler

import (
	"net/http"

	"github.com/mahdi-vajdi/go-image-processor/internal/storage"
)

type publicHandler struct {
	imageStore storage.Storage
}

func NewPublicHandler(imageStore storage.Storage) PublicHandler {
	return &publicHandler{
		imageStore: imageStore,
	}
}

func (p publicHandler) Ping(w http.ResponseWriter, r *http.Request) {
	ResponseJSON(w, http.StatusOK, map[string]string{"message": "pong"})
}
