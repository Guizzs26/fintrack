package rest

import "github.com/go-chi/chi/v5"

func (h *AuthHandler) RegisterRoutes(r chi.Router) {
	r.Route("/auth", func(r chi.Router) {
		r.Post("/signup", h.SignUp)
	})
}
