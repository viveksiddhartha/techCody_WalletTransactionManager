package handlers

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"mfus_WalletTransactionManager/models"
	"net/http"
	"time"

	"github.com/go-playground/validator"
)

// Define middleware to validate requests using the given validator instance
func ValidationMiddleware(validate *validator.Validate, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var requestBody interface{}
		err := json.NewDecoder(r.Body).Decode(&requestBody)
		if err != nil {
			http.Error(w, "Failed to decode JSON", http.StatusBadRequest)
			return
		}

		err = validate.Struct(requestBody)
		if err != nil {
			log.Printf("Validation error: %v\n", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Reset the request body reader
		r.Body = ioutil.NopCloser(bytes.NewBuffer([]byte{}))

		next.ServeHTTP(w, r)
	})
}

// Middleware to validate request body
func ValidateRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			contentType := r.Header.Get("Content-Type")
			if contentType != "application/json" {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(models.ErrorResponse{Message: "Invalid Content-Type. Expected application/json"})
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// Middleware to log requests
func LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.RequestURI, time.Since(start))
	})
}

// Middleware to recover from panics
func RecoverFromPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(models.ErrorResponse{Message: "Internal Server Error"})
			}
		}()
		next.ServeHTTP(w, r)
	})
}
