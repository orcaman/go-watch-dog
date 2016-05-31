package main

import (
	"log"

	"encoding/json"
	"github.com/streamrail/watchdog/models"
	"io/ioutil"
	"path/filepath"
)

type ConfigParser struct {
	Checks []*models.Check `json:"checks,omitempty"`
}

type JSON map[string]interface{}

func NewConfigParser(configPath string) (*ConfigParser, error) {
	rslt := &ConfigParser{Checks: make([]*models.Check, 0)}

	// Scan through the entire config path.
	files, err := ioutil.ReadDir(configPath)
	if err != nil {
		log.Println(JSON{"error": err.Error()})
		return nil, err
	}

	// process each *.config file.
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if filepath.Ext(file.Name()) == ".config" {
			// Read configuration file.
			content, err := ioutil.ReadFile(configPath + file.Name())
			if err != nil {
				log.Println(JSON{"error": err.Error()})
				return nil, err
			}

			config := &ConfigParser{}
			if err = json.Unmarshal(content, &config); err != nil {
				log.Println(JSON{"error": err.Error()})
				return nil, err
			} else {
				// merge config.
				rslt.Checks = append(rslt.Checks, config.Checks...)
			}
		}
	}

	return rslt, nil
}
