package watchers

import (
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/cloud"
	"google.golang.org/cloud/storage"

	"github.com/streamrail/watchdog/models"
)

// func main() {
// 	CheckGCS(&models.Check{})
// }

//Uses c.Query and c.Minimum in config file for test
func CheckGCS(spec *models.Check) error {
	jsonKey, err := ioutil.ReadFile("/etc/service-account.json")
	if err != nil {
		log.Fatal(err)
	}
	conf, err := google.JWTConfigFromJSON(
		jsonKey,
		storage.ScopeFullControl,
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	gcsclnt, err := storage.NewClient(ctx, cloud.WithTokenSource(conf.TokenSource(ctx)))
	if err != nil {
		log.Println(err)
		return nil
	}
	defer gcsclnt.Close()

	buckethandler := gcsclnt.Bucket(spec.GCSBucket)

	q := &storage.Query{
		Prefix: spec.Query,
	}
	objectlist, err := buckethandler.List(ctx, q)
	if err != nil {
		println(err)
		return nil
	}
	for _, val := range objectlist.Results {
		//log.Println(val.Created, val.Name)
		duration := time.Since(val.Created)
		//log.Println("File exist for ", duration.Hours())
		if duration.Hours() > float64(spec.Minimum) {
			return fmt.Errorf("%s was created more than %f hours ago ", val.Name, duration.Hours())
		}
	}

	return nil
}
