package api

import "net/http"

func NewRouter(h *Handler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /payments", h.CreatePayment)
	mux.HandleFunc("GET /payments/{id}", h.GetPayment)
	return withMiddleware(mux)
}
