package main

import (
	"github.com/streamrail/watchdog/models"
	"github.com/streamrail/watchdog/watchers"
	"time"
)

func Schedule(d *watchers.Dispatcher, c *ConfigParser) {
	for _, j := range c.Checks {
		ticker := time.NewTicker(time.Millisecond * time.Duration(j.Frequency))
		go func(check *models.Check) {
			for range ticker.C {
				d.Incoming(check)
			}
		}(j)
	}
}
