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

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		console.Warnf("unable to read request body: %s", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := json.Unmarshal(body, &req); err != nil {
		console.Warnf("unable to parse request body: %s", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	v := s.e.GetVersion(req.VersionID)
	if v == nil {
		console.Warnf("version not found: %s", req.VersionID)
		console.Warnf("this is only populated if the openapi spec is requested :(")
		http.Error(w, "version not found", http.StatusNotFound)
		return
	}

	if err := s.e.LoadVersion(v.imageName(), req.VersionID); err != nil {
		console.Warnf("unable to load version: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id := fmt.Sprintf("%d", rand.Int63())

	response := &Response{
		ID:        id,
		Version:   req.VersionID,
		Input:     req.Input,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
		Logs:      "starting...",
		Source:    "local-api",
		Status:    "starting",
	}

	go s.e.Predict(body, response)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) getPredictions(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")

	response, err := Load(id)
	if err != nil {
		console.Warnf("unable to load prediction: %s", err)
		http.Error(w, "not found", http.StatusNotFound)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
