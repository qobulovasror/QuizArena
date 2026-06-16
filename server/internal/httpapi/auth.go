package httpapi

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"

	"github.com/azizbek12234/quizarena/server/internal/auth"
	"github.com/azizbek12234/quizarena/server/internal/store"
)

type authHandler struct {
	svc      *auth.Service
	validate *validator.Validate
	logger   *slog.Logger
}

type registerReq struct {
	Username string `json:"username" validate:"required,min=3,max=32"`
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=6,max=72"`
}

type loginReq struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type userDTO struct {
	ID       string  `json:"id"`
	Username *string `json:"username"`
	Email    *string `json:"email"`
	IsGuest  bool    `json:"isGuest"`
	Role     string  `json:"role"`
}

type authResp struct {
	Token string  `json:"token"`
	User  userDTO `json:"user"`
}

func toDTO(u store.User) userDTO {
	return userDTO{ID: u.ID.String(), Username: u.Username, Email: u.Email, IsGuest: u.IsGuest, Role: u.Role}
}

func (h *authHandler) guest(w http.ResponseWriter, r *http.Request) {
	res, err := h.svc.Guest(r.Context())
	if err != nil {
		h.logger.Error("guest", "err", err)
		writeErr(w, http.StatusInternalServerError, "mehmon yaratib bo'lmadi")
		return
	}
	writeJSON(w, http.StatusOK, authResp{Token: res.Token, User: toDTO(res.User)})
}

func (h *authHandler) register(w http.ResponseWriter, r *http.Request) {
	var req registerReq
	if !decodeValidate(w, r, h.validate, &req) {
		return
	}
	res, err := h.svc.Register(r.Context(), req.Username, req.Email, req.Password)
	switch {
	case errors.Is(err, auth.ErrEmailTaken):
		writeErr(w, http.StatusConflict, err.Error())
		return
	case err != nil:
		h.logger.Error("register", "err", err)
		writeErr(w, http.StatusInternalServerError, "ro'yxatdan o'tkazib bo'lmadi")
		return
	}
	writeJSON(w, http.StatusCreated, authResp{Token: res.Token, User: toDTO(res.User)})
}

func (h *authHandler) login(w http.ResponseWriter, r *http.Request) {
	var req loginReq
	if !decodeValidate(w, r, h.validate, &req) {
		return
	}
	res, err := h.svc.Login(r.Context(), req.Email, req.Password)
	switch {
	case errors.Is(err, auth.ErrInvalidCreds):
		writeErr(w, http.StatusUnauthorized, err.Error())
		return
	case err != nil:
		h.logger.Error("login", "err", err)
		writeErr(w, http.StatusInternalServerError, "kirib bo'lmadi")
		return
	}
	writeJSON(w, http.StatusOK, authResp{Token: res.Token, User: toDTO(res.User)})
}
