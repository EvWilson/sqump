package main

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

const (
	constUser = "example"
	constPass = "sqump"
)

func main() {
	mux := http.NewServeMux()
	mux.Handle("/", LoggingMiddleware(http.NotFoundHandler().ServeHTTP))
	mux.HandleFunc("/getAuth", LoggingMiddleware(GetAuth))
	mux.HandleFunc("/createThing", LoggingMiddleware(BasicAuthMiddleware(CreateThing)))
	fmt.Println("starting server at 5000")
	err := http.ListenAndServe(":5000", mux)
	if err != nil {
		fmt.Println("error while serving:", err)
		os.Exit(1)
	}
}

func LoggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("incoming: remote: %s, method: %s, path: %s\n", r.RemoteAddr, r.Method, r.URL.RequestURI())
		next.ServeHTTP(w, r)
	})
}

func BasicAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	checkAuth := func(user, pass string) bool {
		usernameHash := sha256.Sum256([]byte(user))
		passwordHash := sha256.Sum256([]byte(pass))
		expectedUsernameHash := sha256.Sum256([]byte(constUser))
		expectedPasswordHash := sha256.Sum256([]byte(constPass))
		usernameMatch := (subtle.ConstantTimeCompare(usernameHash[:], expectedUsernameHash[:]) == 1)
		passwordMatch := (subtle.ConstantTimeCompare(passwordHash[:], expectedPasswordHash[:]) == 1)
		return usernameMatch && passwordMatch
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if ok {
			if !checkAuth(user, pass) {
				next.ServeHTTP(w, r)
				return
			}
		}

		w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})
}

func GetAuth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_, _ = w.Write([]byte("method not allowed"))
		return
	}
	data := []byte(constUser + ":" + constPass)
	b := make([]byte, base64.StdEncoding.EncodedLen(len(data)))
	base64.StdEncoding.Encode(b, data)
	_, _ = w.Write(b)
}

func CreateThing(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("error reading request: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	req := struct {
		Payload string `json:"payload"`
	}{}
	err = json.Unmarshal(b, &req)
	if err != nil {
		http.Error(w, fmt.Sprintf("error unmarshalling request: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	fmt.Println("successfully created print statement for payload:", req.Payload)
}
