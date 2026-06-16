package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// decodeValidate — JSON tanani o'qiydi va struct teglari bo'yicha tekshiradi.
func decodeValidate(w http.ResponseWriter, r *http.Request, v *validator.Validate, dst any) bool {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		writeErr(w, http.StatusBadRequest, "JSON yaroqsiz")
		return false
	}
	if err := v.Struct(dst); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return false
	}
	return true
}
