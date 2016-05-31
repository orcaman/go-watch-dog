package watchers

import (
	"fmt"
	"sync"

	"github.com/streamrail/watchdog/models"
)

type Check *models.CheckState
type CheckMap map[string]Check

type checkCache struct {
	items CheckMap
	sync.RWMutex
}

// retrive an item from cache.
func (c *checkCache) get(key string) (Check, error) {
	c.RLock()
	defer c.RUnlock()

	if item, ok := c.items[key]; ok {
		return item, nil
	} else {
		return item, fmt.Errorf("item missing from cache")
	}
}

// store item in cache
func (c *checkCache) set(key string, value Check) {
	c.Lock()
	defer c.Unlock()

	// make sure item not is missing from cache.
	c.items[key] = value

}
