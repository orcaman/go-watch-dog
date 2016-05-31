package watchers

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/streamrail/watchdog/models"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

func CheckHTTP(spec *models.Check) (int, error) {
	if req, err := http.NewRequest(spec.Request.GetMethod(),
		spec.Request.Url,
		bytes.NewBuffer([]byte(spec.Request.Body))); err != nil {
		return -1, err
	} else {
		if spec.Request.HasAuth() {
			req.SetBasicAuth(spec.Request.Auth.Username, spec.Request.Auth.Password)
		}
		if len(spec.Request.ContentType) > 0 {
			req.Header.Add("Content-Type", spec.Request.ContentType)
		}
		tr := http.Transport{
			Dial: func(network, addr string) (net.Conn, error) {
				timeout := 10 * time.Second
				if spec.IsLatency() {
					if spec.MaxLatencyMs > 0 {
						timeout = time.Duration(spec.MaxLatencyMs) * time.Millisecond
					}
				}
				return net.DialTimeout(network, addr, timeout)
			},
		}

		timeout := 10 * time.Second
		if spec.IsLatency() {
			if spec.MaxLatencyMs > 0 {
				timeout = time.Duration(spec.MaxLatencyMs) * time.Millisecond
			}
		}
		tr.ResponseHeaderTimeout = timeout

		t0 := time.Now()
		if res, err := tr.RoundTrip(req); err != nil {
			return -1, err
		} else {
			defer res.Body.Close()
			t1 := time.Now()
			latency := t1.Sub(t0)
			err := ParseHTTPResultPerSpec(spec, res)
			if err != nil {
				return -1, err
			}
			return int(latency.Seconds() * 1000), nil
		}
	}
}

func ParseHTTPResultPerSpec(spec *models.Check, res *http.Response) error {
	if spec.IsLatency() {
		if spec.Statuscode > 0 {
			if res.StatusCode == spec.Statuscode {
				return nil
			} else {
				return fmt.Errorf("error: expected statuscode to be %d but got %d instead", spec.Statuscode, res.StatusCode)
			}
		} else {
			return nil
		}
	}

	if spec.IsStatuscode() {
		if res.StatusCode != spec.Statuscode {
			return fmt.Errorf("error: expected statuscode to be %d but got %d instead", spec.Statuscode, res.StatusCode)
		} else {
			return nil
		}
	}

	strResp, _ := ioutil.ReadAll(res.Body)

	if spec.IsChange() {
		newMd5 := GetMD5Hash(strResp)
		cacheKey := spec.Request.Key()
		item, err := cache.get(cacheKey)
		if err != nil { // cache miss
			item.MD5 = newMd5
			// fmt.Printf("initializng content with md5: %s\n", newMd5)
			return nil
		} else {
			if len(item.MD5) > 0 {
				if item.MD5 == newMd5 {
					return nil
				} else {
					item.MD5 = newMd5
					return fmt.Errorf("error: expected md5 to be %d but got %d instead", item, newMd5)
				}
			}
		}
	}

	if spec.IsContains() {
		if strings.Contains(string(strResp), spec.Contains) {
			return nil
		} else {
			return fmt.Errorf("error: expected response to %s to contains string %s", spec.Request.Url, spec.Contains)
		}
	}
	return nil
}

func GetMD5Hash(text []byte) string {
	hash := md5.Sum(text)
	return hex.EncodeToString(hash[:])
}
