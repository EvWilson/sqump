package util

import (
	"net/http"

	"github.com/gofrs/uuid/v5"
)

const (
	IDCookieName  = "sqump-id"
	ErrCookieName = "sqump-error"
)

func GetID(r *http.Request) (uuid.UUID, error) {
	cookie, err := r.Cookie(IDCookieName)
	if err != nil {
		return uuid.Nil, err
	}
	return uuid.FromString(cookie.Value)
}

func SetNewID(w http.ResponseWriter) error {
	uid, err := uuid.NewV4()
	if err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     IDCookieName,
		Value:    uid.String(),
		Path:     "/",
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
	})
	return nil
}

func SetErrorCookie(w http.ResponseWriter, err string) {
	http.SetCookie(w, &http.Cookie{
		Name:     ErrCookieName,
		Value:    err,
		Path:     "/",
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
	})
}

func GetErrorOnRequest(w http.ResponseWriter, req *http.Request) string {
	cookie, err := req.Cookie(ErrCookieName)
	if err == nil {
		http.SetCookie(w, &http.Cookie{
			Name:     ErrCookieName,
			Value:    "",
			MaxAge:   -1,
			Secure:   false,
			SameSite: http.SameSiteStrictMode,
		})
		return cookie.Value
	} else {
		return ""
	}
}
