package watchers

//package watchers

import (
	"fmt"
	"io/ioutil"
	"strings"
	"sync"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"

	"github.com/streamrail/watchdog/models"
)

var (
	svc                  *compute.Service
	GCEmonitoredServices []*models.Server
	jwtConfig            = "/etc/sr.json"
	projectID            = "streamrail"
)

//JUST FOR DEBUG :)
// func main() {
// 	fmt.Println("Start...")
// 	time.Sleep(1 * time.Minute)
// }

func init() {

	pemKeyBytes, err := ioutil.ReadFile(jwtConfig)
	if err != nil {
		fmt.Println("Error")
		panic(err)
	}
	t, err := google.JWTConfigFromJSON(
		pemKeyBytes,
		"https://www.googleapis.com/auth/compute",
		"https://www.googleapis.com/auth/cloud-platform")
	//t := jwt.NewToken(c.accountEmailAddress, bigquery.BigqueryScope, pemKeyBytes)
	client := t.Client(oauth2.NoContext)

	svc, err = compute.New(client)
	if err != nil {
		//return nil, err
		fmt.Println("Error")
		panic("Can't connect to gcloud compute api")
	}

	ticker := time.NewTicker(time.Minute * 25)
	go func() {
		for {
			GCEmonitoredServices = listAllGCEServer()
			<-ticker.C
		}
	}()

}

func GCECheckDeployHash(c *models.Check) (string, error) {

	if GCEmonitoredServices == nil && len(GCEmonitoredServices) == 0 {
		return "", nil
	}

	randDeployHash := ""
	for _, s := range GCEmonitoredServices {
		//Only search for servers with the relevant tag
		//fmt.Println("tags", s.Tags, "tag", c.GCEInstanceTag)
		if contains(s.Tags, c.GCEInstanceTag) {
			randDeployHash = s.DeployHash

			if s.DeployHash != randDeployHash {
				return randDeployHash, fmt.Errorf("non-identical deploy hash was found in gce for instance group tag: %s with deployhash %s", c.GCEInstanceTag, s.DeployHash)
			}
		}
	}
	return randDeployHash, nil
}

func listAllGCEServer() []*models.Server {

	fmt.Println("in listAllGCEServer()")

	var wg sync.WaitGroup
	rslt := []*models.Server{}

	computesvc := compute.NewInstancesService(svc)
	zones, err := getAllZonesWithPrefix("us-central1")
	if err != nil {
		fmt.Println(err)
	}
	for _, z := range zones {
		instancelistcall := computesvc.List(projectID, z)
		for {
			instancelist, err := instancelistcall.Do()
			if err != nil {
				fmt.Println("Error in listing gce instances ", err)
				break
			}
			for _, v := range instancelist.Items {

				wg.Add(1)
				go func(v *compute.Instance) {
					defer wg.Done()
					publicIP := v.NetworkInterfaces[0].AccessConfigs[0].NatIP
					//dnsName := *inst.PublicDnsName
					//getDeployHash(publicIP)

					rslt = append(rslt, &models.Server{
						Name:          v.Name,
						PublicDnsName: publicIP,
						DeployHash:    getDeployHash(publicIP),
						Tags:          v.Tags.Items,
					})

				}(v)
			}
			if instancelist.NextPageToken != "" {
				instancelistcall.PageToken(instancelist.NextPageToken)
				continue
			} else {
				break
			}

		}

	}
	wg.Wait()

	//fmt.Println("Returning", rslt)
	return rslt

}

func getAllZonesWithPrefix(prefix string) ([]string, error) {
	//fmt.Println("in getAllZonesWithPrefix")
	var zones []string
	zonessvc := compute.NewZonesService(svc)
	zonelistcall := zonessvc.List(projectID)

	for {
		zonelist, err := zonelistcall.Do()
		if err != nil {
			return nil, fmt.Errorf("Zonelist Error", err)
		}

		for _, v := range zonelist.Items {
			// fmt.Println("zone", i)
			// fmt.Println("Zone:", v.Name)
			if strings.HasPrefix(v.Name, prefix) {

				zones = append(zones, v.Name)
			}

		}

		if zonelist.NextPageToken != "" {
			zonelistcall.PageToken(zonelist.NextPageToken)
			continue
		}
		return zones, nil
	}

}

func contains(slc []string, search string) bool {
	for _, item := range slc {
		if item == search {
			return true
		}
	}
	return false
}
