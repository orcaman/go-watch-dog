package models

import (
	"encoding/json"
	"strings"
)

type Check struct {
	Type           string       `json:"type,omitempty"`
	Name           string       `json:"name,omitempty"`
	Contains       string       `json:"contains,omitempty"`
	InstanceGroup  string       `json:"instanceGroup,omitempty"`
	GCEInstanceTag string       `json:"gceInstanceTag,omitempty"`
	Minimum        int          `json:"minimum,omitempty"`
	Request        *HttpRequest `json:"request,omitempty"`
	Frequency      int          `json:"frequency,omitempty"`
	Statuscode     int          `json:"statuscode,omitempty"`
	MaxLatencyMs   int          `json:"maxLatencyMs,omitempty"`
	AlertAfter     int          `json:"alertafter"`
	Mailto         string       `json:"mailto,omitempty"`
	GCSBucket      string       `json:"gcsbucket,omitempty"`

	Query string `json:"query,omitempty"`

	SlackWebHookUrl    string `json:"slackWebHook,omitempty"`
	SessionID          float64
	Alert              bool   `json:"alert,omitempty"`
	MeasuredLatency    int    `json:"measuredLatencyMs,omitempty"`
	MeasuredDeployHash string `json:"deployHash"`
}

type CheckState struct {
	MD5        string
	NumOfFails int
	Alert      bool
}

func NewCheckState(md5 string, alert bool) *CheckState {
	return &CheckState{
		MD5:        md5,
		Alert:      alert,
		NumOfFails: 0,
	}
}

func (c *Check) Key() string {
	if c.Request != nil {
		return c.Type + c.Name + c.InstanceGroup + c.Request.Key()
	}
	return c.Type + c.Name + c.InstanceGroup
}

func (c *Check) IsChange() bool {
	return strings.ToLower(c.Type) == "change"
}

func (c *Check) IsBqCount() bool {
	return strings.ToLower(c.Type) == "bqcount"
}

func (c *Check) IsLatency() bool {
	return strings.ToLower(c.Type) == "latency"
}

func (c *Check) IsStatuscode() bool {
	return strings.ToLower(c.Type) == "statuscode"
}

func (c *Check) IsDeployHash() bool {
	return strings.ToLower(c.Type) == "deployhash"
}

func (c *Check) IsGCEDeployHash() bool {
	return strings.ToLower(c.Type) == "gcedeployhash"
}

func (c *Check) IsGCS() bool {
	return strings.ToLower(c.Type) == "gcs"
}

func (c *Check) IsMinInstanceCount() bool {
	return strings.ToLower(c.Type) == "mininstancecount"
}

func (c *Check) IsContains() bool {
	return strings.ToLower(c.Type) == "contains"
}

func (c *Check) ToJsonString() string {
	str, _ := json.Marshal(c)
	return string(str)
}
