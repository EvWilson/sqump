package util

import "net/http"

func SetErrorCookie(w http.ResponseWriter, err string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "error",
		Value:    err,
		Path:     "/",
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
	})
}

func GetErrorOnRequest(w http.ResponseWriter, req *http.Request) string {
	cookie, err := req.Cookie("error")
	if err == nil {
		http.SetCookie(w, &http.Cookie{
			Name:     "error",
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
