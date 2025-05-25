package rest

import (
	"encoding/json"
	"net/http"
)

type AuthHandler struct {
}

func NewAuthHandler() AuthHandler {
	return AuthHandler{}
}

func (h *AuthHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var req SignUpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		msg := NewBadRequestError("invalid request body")
		json.NewEncoder(w).Encode(msg)
		return
	}

	if httpErr := ValidateHttpData(req); httpErr != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(httpErr.Code)
		json.NewEncoder(w).Encode(httpErr)
		return
	}
}
