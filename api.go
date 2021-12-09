package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type APIResponse struct {
	Status  string      `json:"status"`
	Errors  Errors      `json:"errors"`
	Error   StringError `json:"error"`
	Message string
}
type StringError string

func (s StringError) Error() string {
	return string(s)
}

type Errors map[string][]string

const (
	StatusSuccess string = "success"
	StatusFailure string = "failure"
)

// ParseAPIResponse unmarshals the response and closes the reader
func ParseAPIResponse(httpResponse *http.Response) (*APIResponse, error) {
	defer httpResponse.Body.Close()
	decoder := json.NewDecoder(httpResponse.Body)
	response := APIResponse{}
	err := decoder.Decode(&response)
	if err != nil {
		return nil, fmt.Errorf("parse api response: %w", err)
	}
	return &response, nil
}
