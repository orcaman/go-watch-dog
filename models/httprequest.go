package models

import (
	"encoding/json"
	"fmt"
	"strings"
)

type HttpRequest struct {
	Url         string    `json:"url,omitempty"`
	Method      string    `json:"method,omitempty"`
	Body        string    `json:"body,omitempty"`
	Auth        *HttpAuth `json:"auth,omitempty"`
	ContentType string    `json:"contentType,omitempty"`
}

// unique identifier for this reqeust
func (h *HttpRequest) Key() string {
	sBody, _ := json.Marshal(h.Body)
	sAuth := []byte("")
	if h.HasAuth() {
		sAuth, _ = json.Marshal(h.Auth)
	}
	return fmt.Sprintf("%s_%s_%s_%s", h.Url, h.Method, sBody, sAuth)
}

func (h *HttpRequest) IsPost() bool {
	return h.GetMethod() == "POST"
}

func (h *HttpRequest) GetMethod() string {
	return strings.ToUpper(h.Method)
}

func (h *HttpRequest) HasAuth() bool {
	return h.Auth != nil && len(h.Auth.Password) > 0 && len(h.Auth.Username) > 0
}
