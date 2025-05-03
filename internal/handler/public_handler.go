package handler

import (
	"net/http"
)

type publicHandler struct {
}

func NewPublicHandler() PublicHandler {
	return &publicHandler{}
}

func (p publicHandler) Ping(w http.ResponseWriter, r *http.Request) {
	ResponseJSON(w, http.StatusOK, map[string]string{"message": "pong"})
}
