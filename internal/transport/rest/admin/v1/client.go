package v1

import (
	"encoding/json"
	"net/http"

	"github.com/tuanta7/oauth-server/internal/interactors/client"
)

type ClientHandler struct {
	clientUC client.UseCase
}

func NewClientHandler(clientUC client.UseCase) *ClientHandler {
	return &ClientHandler{
		clientUC: clientUC,
	}
}

func (h *ClientHandler) List(w http.ResponseWriter, r *http.Request) {
	clients, err := h.clientUC.List(r.Context(), 1, 10)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	jsonPayload, err := json.Marshal(clients)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(jsonPayload)
	return
}
