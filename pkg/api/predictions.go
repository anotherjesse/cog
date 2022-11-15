package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

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

func (r *Response) Save() error {
	content, err := json.MarshalIndent(r, "", " ")
	if err != nil {
		return fmt.Errorf("unable to serialize response to json: %w", err)
	}

	filename := fmt.Sprintf("predictions/%s.json", r.ID)
	return ioutil.WriteFile(filename, content, 0o644)
}

func Load(id string) (*Response, error) {
	filename := fmt.Sprintf("predictions/%s.json", id)
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("unable to read response file: %w", err)
	}

	r := Response{}

	err = json.Unmarshal(content, &r)
	if err != nil {
		return nil, fmt.Errorf("unable to parse response file: %w", err)
	}

	return &r, nil
}
