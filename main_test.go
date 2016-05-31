package main

import (
	"testing"
)

func TestFoo(t *testing.T) {
	_, err := NewConfigParser(configPath)
	if err != nil {
		t.Error(err)
	}
}
