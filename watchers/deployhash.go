package watchers

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/streamrail/watchdog/models"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

var monitoredServices map[string]*models.MonitoredService

func InitDeployHashes() {
	fmt.Print("initializing monitored services, please wait... ")
	monitoredServices = getMonitoredServices()
	fmt.Println("done.")
	ticker := time.NewTicker(time.Minute * 5)
	go func() {
		for range ticker.C {
			monitoredServices = getMonitoredServices()
		}
	}()
}

func main() {
	rslt := getMonitoredServices()
	j, _ := json.Marshal(rslt)
	fmt.Println(string(j))
}

func CheckDeployHash(c *models.Check) (string, error) {
	if monitoredServices == nil {
		return "", nil
	}
	service := monitoredServices[c.InstanceGroup]
	if service == nil {
		return "", nil
	}
	servers := service.Servers
	if servers == nil {
		return "", nil
	}

	if len(servers) == 0 {
		return "", nil
	}
	randomServer := ""
	for k, _ := range servers {
		randomServer = k
		break
	}
	deployHash := servers[randomServer].DeployHash
	for _, s := range servers {
		if s.DeployHash != deployHash {
			return deployHash, fmt.Errorf("non-identical deploy hash was found in the same instance group: %s", c.InstanceGroup)
		}
	}
	return deployHash, nil
}

func CheckMinInstanceCount(c *models.Check) error {
	if monitoredServices == nil {
		return nil
	}
	service := monitoredServices[c.InstanceGroup]
	if service == nil {
		return fmt.Errorf("monitoredServices not ready")
	}
	servers := service.Servers
	if servers == nil {
		return fmt.Errorf("InstanceGroup not found: %s", c.InstanceGroup)
	}

	if len(servers) < c.Minimum {
		return fmt.Errorf("instance group %s: expected at least %d servers but found only %d", c.InstanceGroup, c.Minimum, len(servers))
	}
	return nil
}

func getMonitoredServices() map[string]*models.MonitoredService {
	rslt := make(map[string]*models.MonitoredService)
	svc := ec2.New(session.New(), &aws.Config{Region: aws.String("us-east-1")})

	resp, err := svc.DescribeInstances(nil)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}

	for idx, _ := range resp.Reservations {
		for _, inst := range resp.Reservations[idx].Instances {
			for _, t := range inst.Tags {
				if *t.Key == "Name" {
					name := *t.Value
					dnsName := *inst.PublicDnsName
					if rslt[name] == nil {
						rslt[name] = &models.MonitoredService{
							Name: name,
							Servers: map[string]*models.Server{
								dnsName: &models.Server{
									PublicDnsName: dnsName,
									DeployHash:    getDeployHash(dnsName),
								},
							},
						}
					} else {
						if rslt[name].Servers == nil {
							rslt[name].Servers = map[string]*models.Server{
								dnsName: &models.Server{
									PublicDnsName: dnsName,
									DeployHash:    getDeployHash(dnsName),
								},
							}
						} else {
							rslt[name].Servers[dnsName] = &models.Server{
								PublicDnsName: dnsName,
								DeployHash:    getDeployHash(dnsName),
							}
						}
					}
				}
			}
		}
	}

	return rslt
}

func getExpVarResponse(url string, username string, password string) map[string]interface{} {
	req, _ := http.NewRequest("GET", url, nil)
	if len(username) > 0 && len(password) > 0 {
		req.SetBasicAuth(username, password)
	}
	tr := http.Transport{
		Dial: func(network, addr string) (net.Conn, error) {
			timeout := 2 * time.Second
			return net.DialTimeout(network, addr, timeout)
		},
	}
	fmt.Println("url: " + url)
	tr.ResponseHeaderTimeout = 2 * time.Second
	res, err := tr.RoundTrip(req)
	if err == nil && res.StatusCode == 200 {
		str, _ := ioutil.ReadAll(res.Body)
		defer res.Body.Close()
		var obj map[string]interface{}
		json.Unmarshal(str, &obj)
		return obj
	}
	return nil
}

func getDeployHash(publicDnsName string) string {
	if len(publicDnsName) == 0 {
		return ""
	}

	url := "http://" + publicDnsName + "/debug/vars"
	obj := getExpVarResponse(url, "", "")
	if obj == nil {
		url = "http://" + publicDnsName + "/expose"
		obj = getExpVarResponse(url, "streamrail", "fireintheh0le")
	}
	if obj != nil && obj["DeployHashKey"] != nil {
		hash := obj["DeployHashKey"].(string)
		return hash
	}
	return ""
}
