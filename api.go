package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

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
    router.HandleFunc("/account/{id}", makeHTTPHandler(s.handleGetAccount)).Methods("GET")
    router.HandleFunc("/account/{id}", makeHTTPHandler(s.handleDeleteAccount)).Methods("DELETE")
    router.HandleFunc("/account/{id}/balance", makeHTTPHandler(s.handleBalance)).Methods("PATCH")

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

func SaveAccount(account *Account) error {
    if err := os.MkdirAll("db", os.ModePerm); err != nil {
        return err
    }

    filePath := filepath.Join("db", fmt.Sprintf("%d.json", account.ID))
    file, err := os.Create(filePath)

    if err != nil {
        return err
    }
    defer file.Close()

    return json.NewEncoder(file).Encode(account)
}

const idFilePath = "db/last_id.txt"

func GetNextID() (int, error) {
    var lastID int

    if _, err := os.Stat(idFilePath); os.IsNotExist(err){
        if err := os.WriteFile(idFilePath, []byte("0"), 0644); err != nil {
            return 0, fmt.Errorf("could not create last_id.txt: %w", err)
        }
        lastID = 0
    } else {
        data, err := os.ReadFile(idFilePath)
        if err != nil {
            return 0, fmt.Errorf("could not read last_id.txt: %w", err)
        }
        if _, err := fmt.Sscanf(string(data), "%d", &lastID); err != nil {
            return 0, fmt.Errorf("could not parse lastID: %w", err)
        }
    }

    nextID := lastID + 1

    if err := os.WriteFile(idFilePath, []byte(fmt.Sprintf("%d", nextID)), 0644); err != nil {
        return 0, fmt.Errorf("could not write to last_id.txt %w", err)
    }

    return nextID, nil
}

// HELPERS

func (s *APIServer) handleAccount(w http.ResponseWriter, r *http.Request) error {
    if r.Method == "GET" {
        return s.handleGetAccount(w, r)
    }
    if r.Method == "POST" {
        return s.handleCreateAccount(w)
    }
    if r.Method == "DELETE" {
        return s.handleDeleteAccount(w, r)
    }

    return fmt.Errorf("method not allowed %s", r.Method)
}

func (s *APIServer) handleBalance(w http.ResponseWriter, r *http.Request) error {
    vars := mux.Vars(r)

    idStr := vars["id"]

    var id int
    _, err := fmt.Sscanf(idStr, "%d", &id)
    if err != nil {
        return fmt.Errorf("invalid account ID: %s", idStr)
    }

    filePath := filepath.Join("db", fmt.Sprintf("%d.json", id))

    data, err := os.ReadFile(filePath)
    if os.IsNotExist(err) {
        return fmt.Errorf("account not found: %d", id)
    } else if err != nil {
        return fmt.Errorf("could not read account file: %w", err)
    }

    var account Account
    if err := json.Unmarshal(data, &account); err != nil {
        return fmt.Errorf("could not decode account: %w", err)
    }

    var requestBody struct {
        Amount int `json:"amount"`
    }

    if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
            return fmt.Errorf("could not decode request body: %w", err)
    }

    account.Balance += int64(requestBody.Amount)
    if err := SaveAccount(&account); err != nil {
        return fmt.Errorf("could not update account_%d: %w", account.ID, err)
    }

    log.Printf("added %d to account_%d successfully", requestBody.Amount, id)

    return WriteJSON(w, http.StatusOK, account)
}


func (s *APIServer) handleGetAccount(w http.ResponseWriter, r *http.Request) error {
    vars := mux.Vars(r)

    idStr := vars["id"]

    var id int
    _, err := fmt.Sscanf(idStr, "%d", &id)
    if err != nil {
        return fmt.Errorf("invalid account ID: %s", idStr)
    }

    filePath := filepath.Join("db", fmt.Sprintf("%d.json", id))

    data, err := os.ReadFile(filePath)
    if os.IsNotExist(err) {
        return fmt.Errorf("account not found: %d", id)
    } else if err != nil {
        return fmt.Errorf("could not read account file: %w", err)
    }

    var account Account
    if err := json.Unmarshal(data, &account); err != nil {
        return fmt.Errorf("could not decode account: %w", err)
    }

    return WriteJSON(w, http.StatusOK, account)
}

func (s *APIServer) handleCreateAccount(w http.ResponseWriter) error {
    id, err := GetNextID()
    account := NewAccount(id, "BOOM", "Baby")

    log.Println(account)

    if err != nil {
        return fmt.Errorf("could not get next id: %v", err)
    }

    log.Printf("next id: %d", id)

    if err := SaveAccount(account); err != nil {
        return fmt.Errorf("couldn't save account: %w", err)
    }

    return WriteJSON(w, http.StatusOK, account)
}

func (s *APIServer) handleDeleteAccount(w http.ResponseWriter, r *http.Request) error {
    vars := mux.Vars(r)

    idStr := vars["id"]

    var id int
    _, err := fmt.Sscanf(idStr, "%d", &id)
    if err != nil {
        return fmt.Errorf("invalid account ID: %s", idStr)
    }

    filePath := filepath.Join("db", fmt.Sprintf("%d.json", id))

    if _, err := os.Stat(filePath); os.IsNotExist(err) {
        return fmt.Errorf("account not found: %w", err)
    } else if  err != nil {
        return fmt.Errorf("could not read file: %w", err)
    }

    if err := os.Remove(filePath); err != nil {
        return fmt.Errorf("could not remove file: %w", err)
    }

    return WriteJSON(w, http.StatusOK, ApiError{Error: fmt.Sprintf("account %d deleted successfully", id)})
}
