package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

//TYPES

type APIServer struct {
    listenAddr string
}

type apiFunc func(w http.ResponseWriter, r *http.Request) error

type ApiError struct {
    Error string `json:"error"`
}

// RANDOM STUFF

func WriteJSON(w http.ResponseWriter, status int, v any) error {
    w.WriteHeader(status)
    w.Header().Set("Content-Type", "application/json")
    return json.NewEncoder(w).Encode(v)
}

func NewServer(listenAddr string) *APIServer {
    return &APIServer{
        listenAddr: listenAddr,
    }
}

func (s *APIServer) Run() {
    router := mux.NewRouter()

    router.HandleFunc("/account", makeHTTPHandler(s.handleAccount))

    log.Println("api server listening on -> ", s.listenAddr)

    if err := http.ListenAndServe(s.listenAddr, router); err != nil {
        log.Fatalf("cound not start server: %s\n", err)
    }
}

func makeHTTPHandler(f apiFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if err := f(w, r); err != nil {
            WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
        }
    }
}

// HELPERS

func (s *APIServer) handleAccount(w http.ResponseWriter, r *http.Request) error {
    if r.Method == "GET" {
        return s.handleGetAccount(w, r)
    }
    if r.Method == "POST" {
        return s.handleCreateAccount(w, r)
    }
    if r.Method == "DELETE" {
        return s.handleDeleteAccount(w, r)
    }

    return fmt.Errorf("method not allowed %s", r.Method)
}

func (s *APIServer) handleBalance(w http.ResponseWriter, r *http.Request) error {
    return nil
}


func (s *APIServer) handleGetAccount(w http.ResponseWriter, r *http.Request) error {
    return nil
}

func (s *APIServer) handleCreateAccount(w http.ResponseWriter, r *http.Request) error {
    account := NewAccount("BOOM", "Baby")

    log.Println(account)

    return WriteJSON(w, http.StatusOK, account)
}

func (s *APIServer) handleDeleteAccount(w http.ResponseWriter, r *http.Request) error {
    return nil
}
