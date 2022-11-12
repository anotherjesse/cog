package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/replicate/cog/pkg/util/console"
)

func replicateRequest(authorization string, path string) (string, error) {
	url := "https://api.replicate.com" + path

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", authorization)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

type VersionsResponse struct {
	Previous string `json:"previous"`
	Next     string `json:"next"`
	Results  []struct {
		ID         string `json:"id"`
		CreatedAt  string `json:"created_at"`
		CogVersion string `json:"cog_version"`
	}
}

type Version struct {
	userName    string
	modelName   string
	versionID   string
	openapiSpec string
}

func (v *Version) imageName() string {
	return fmt.Sprintf("%s/%s/%s@sha256:%s",
		"r8.im", v.userName, v.modelName, v.versionID)
}

func (e *Engine) ensureVersion(versionID string, userName string, modelName string, authorization string) error {
	if !e.HasVersion(versionID) {

		path := fmt.Sprintf("/v1/models/%s/%s/versions/%s", userName, modelName, versionID)

		console.Infof("getting openapi spec from replicate: %s", path)
		spec, err := replicateRequest(authorization, path)
		if err != nil {
			return err
		}
		v := Version{
			userName:    userName,
			modelName:   modelName,
			versionID:   versionID,
			openapiSpec: spec,
		}
		e.AddVersion(v)
	}
	return nil
}

func (s *Server) modelOpenAPISpec(w http.ResponseWriter, r *http.Request) {

	versionID := chi.URLParam(r, "versionId")
	userName := chi.URLParam(r, "userName")
	modelName := chi.URLParam(r, "modelName")
	authorization := r.Header.Get("Authorization")

	err := s.e.ensureVersion(versionID, userName, modelName, authorization)
	if err != nil {
		console.Warnf("unable to ensure version: %s", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	v := s.e.GetVersion(versionID)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(v.openapiSpec))
}

func (s *Server) modelVersions(w http.ResponseWriter, r *http.Request) {

	path := r.URL.Path
	userName := chi.URLParam(r, "userName")
	modelName := chi.URLParam(r, "modelName")
	authorization := r.Header.Get("Authorization")

	console.Infof("getting model versions from replicate: %s", path)
	versions, err := replicateRequest(authorization, path)
	if err != nil {
		console.Warnf("unable to get model versions from replicate: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	versionsResponse := VersionsResponse{}
	err = json.Unmarshal([]byte(versions), &versionsResponse)
	if err != nil {
		console.Warnf("unable to unmarshal versions: %s", err)
	} else {
		// FIXME(ja): the openapi spec is actually in the response
		// but I don't know how to grab it as a string (or should I?)
		for _, result := range versionsResponse.Results {
			err = s.e.ensureVersion(result.ID, userName, modelName, authorization)
			if err != nil {
				console.Warnf("unable to ensure version: %s", err)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(versions))
}
