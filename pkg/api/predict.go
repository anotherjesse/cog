package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/replicate/cog/pkg/docker"
	"github.com/replicate/cog/pkg/util/console"
)

// json schema for request body
type Request struct {
	Input     Inputs `json:"input"`
	VersionID string `json:"version"`
}

type Inputs map[string]interface{}

func ensureImageExists(imageName string) error {
	exists, err := docker.ImageExists(imageName)
	if err != nil {
		return fmt.Errorf("Failed to determine if %s exists: %w", imageName, err)
	}
	if !exists {
		console.Infof("Pulling image: %s", imageName)
		if err := docker.Pull(imageName); err != nil {
			return fmt.Errorf("Failed to pull %s: %w", imageName, err)
		}
	}
	return err
}

func (s *Server) predictAPI(w http.ResponseWriter, r *http.Request) {
	var req Request
	body, _ := ioutil.ReadAll(r.Body)
	fmt.Println("body", string(body))

	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	v := s.e.GetVersion(req.VersionID)
	fmt.Println("input", req.Input)
	if v == nil {
		fmt.Println("version not found", req.VersionID)
		fmt.Println("this is only populated if the openapi spec is requested :(")
		http.Error(w, "version not found", http.StatusNotFound)
		return
	}
	imageName := v.imageName

	if err := s.e.LoadVersion(imageName, req.VersionID); err != nil {
		fmt.Print("error loading version", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id := fmt.Sprintf("%d", rand.Int63())

	s.e.result = &Response{
		ID:        id,
		Version:   req.VersionID,
		Input:     req.Input,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
		Logs:      "starting...",
		Source:    "local-api",
		Status:    "starting",
	}

	go s.e.Predict(body)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(s.e.result); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

type Response struct {
	ID          string       `json:"id"`
	Version     string       `json:"version"`
	URLs        URLs         `json:"urls"`
	CreatedAt   string       `json:"created_at"`
	CompletedAt string       `json:"completed_at"`
	Source      string       `json:"source"`
	Status      string       `json:"status"`
	Input       Inputs       `json:"input"`
	Output      *interface{} `json:"output"`
	Error       string       `json:"error"`
	Logs        string       `json:"logs"`
}

type URLs struct {
	Get    string `json:"get"`
	Cancel string `json:"cancel"`
}

type Output struct {
	Text string `json:"text"`
}

func (s *Server) getPredictions(w http.ResponseWriter, r *http.Request) {
	if s.e.result == nil {
		http.Error(w, "No result", http.StatusNotFound)
		return
	}

	if s.e.result.ID != chi.URLParam(r, "id") {
		http.Error(w, "No result", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(s.e.result); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
