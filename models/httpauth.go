package models

type HttpAuth struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}
