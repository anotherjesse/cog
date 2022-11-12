package api

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func replicate(authorization string, path string) (string, error) {
	url := "https://api.replicate.com" + path
	// create a new request to replicate using Authorization header
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

type Version struct {
	userName    string
	modelName   string
	versionID   string
	imageName   string
	openapiSpec string
}

func (s *Server) modelOpenAPISpec(w http.ResponseWriter, r *http.Request) {

	versionID := chi.URLParam(r, "versionId")

	if s.e.HasVersion(versionID) == false {
		authorization := r.Header.Get("Authorization")
		path := r.URL.Path

		userName := chi.URLParam(r, "userName")
		modelName := chi.URLParam(r, "modelName")

		imageName := fmt.Sprintf("%s/%s/%s@sha256:%s", "r8.im", userName, modelName, versionID)

		spec, err := replicate(authorization, path)
		if err != nil {
			fmt.Println("error getting spec", path)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		v := Version{
			userName:    userName,
			modelName:   modelName,
			versionID:   versionID,
			imageName:   imageName,
			openapiSpec: spec,
		}
		s.e.AddVersion(v)
	}

	v := s.e.GetVersion(versionID)

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(v.openapiSpec))
}
