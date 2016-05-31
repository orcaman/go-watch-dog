package watchers

import (
	"encoding/json"
	"github.com/streamrail/watchdog/models"
	"math/rand"
	"time"
)

var (
	reportsValue = make(map[string]*models.Check)
	cache        = checkCache{
		items: make(map[string]Check),
	}
)

type Dispatcher struct {
	incoming chan *models.Check
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		incoming: make(chan *models.Check),
	}
}

func (r *Dispatcher) Listen() {
	go func() {
		for c := range r.incoming {
			c.SessionID = rand.Float64()
			go func(c *models.Check) {
				key := c.Key()
				latency := -1
				deployHash := ""
				var tmp Check

				var err error
				if c.IsStatuscode() || c.IsLatency() || c.IsChange() || c.IsContains() {
					latency, err = CheckHTTP(c)
				}
				if c.IsDeployHash() {
					deployHash, err = CheckDeployHash(c)
					c.MeasuredDeployHash = deployHash
				}
				if c.IsGCEDeployHash() {
					deployHash, err = GCECheckDeployHash(c)
					c.MeasuredDeployHash = deployHash
				}
				if c.IsMinInstanceCount() {
					err = CheckMinInstanceCount(c)
				}
				if c.IsBqCount() {
					err = CheckMinBQCount(c.Query, c.Minimum)
				}
				if c.IsGCS() {
					err = CheckGCS(c)
				}

				// try to get check state from cache

				//tmp := Cache[key]
				tmp, cacheErr := cache.get(key)
				// log.Println(tmp, cacheErr, cache)
				if err != nil { // there was a "real" problem with the check
					if cacheErr == nil { // if this check existed in the cache already
						//log.Println("failed but IN in cache")
						if !tmp.Alert { // ONLY if it was not alerted, send alert
							tmp.NumOfFails++
							//	log.Println("Alert (Cached): ", c.Name, "  - NumofFails: ", tmp.NumOfFails, "Alert After is :", c.AlertAfter)
							if tmp.NumOfFails >= c.AlertAfter {
								c.Alert = true // set check alert to true
								tmp.Alert = true
								SendNotification(c, err)
							}
						}
						// otherwise - do nothing (check exists in cache already and it's alerted)
					} else { // check didnt exist in cache - alert and put in cache (bottom of func)
						//log.Println("failed but NOT in cache")
						tmp = models.NewCheckState("", false) //Only change alert if AlertAfter is passed
						tmp.NumOfFails++
						//log.Println("Alert: ", c.Name, "  NumOFails: ", tmp.NumOfFails, "Alert After is :", c.AlertAfter)
						if tmp.NumOfFails >= c.AlertAfter {
							c.Alert = true
							tmp.Alert = true
							SendNotification(c, err)
						}
					}
				} else { // problem was sorted or there is no problem at all
					//log.Println("Check passed", c.Name)
					if latency != -1 {
						c.MeasuredLatency = latency
					}
					c.Alert = false // set alert to false
					if tmp != nil { // if there is an entry in the cache
						if tmp.Alert { // and it had an error
							tmp.Alert = false // update cache entry
							tmp.NumOfFails = 0
							SendNotification(c, nil) // send "all good" email
						}

					} else { //prepare entry for cache
						tmp = models.NewCheckState("", false)
					}

				}
				//log.Println("cacheing ", tmp, "With Key ", key)
				//Cache[key] = tmp         // update cache state
				cache.set(key, tmp)      // update cache state
				reportsValue[c.Name] = c // exp var reporter
				//log.Println("finished here")
			}(c)
		}
	}()
}

func (r *Dispatcher) Incoming(c *models.Check) {
	r.incoming <- c
}

func (r *Dispatcher) GetReportsValueJson() []byte {
	asArray := []map[string]interface{}{}
	for i, j := range reportsValue {
		tmp := make(map[string]interface{})
		tmp["title"] = i
		tmp["desc"] = j
		asArray = append(asArray, tmp)
	}
	js, _ := json.Marshal(asArray)
	return js
}
