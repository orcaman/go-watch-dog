package watchers

import (
	"fmt"
	"os"
	//"github.com/Dailyburn/bigquery/client"
	"github.com/streamrail/bigquery/client"
	"strconv"
)

const PEM_PATH = "/etc/service-account.json"
const PROJECTID = os.Getenv("WATCHDOG_BQPROJECT")
const DATASET = os.Getenv("WATCHDOG_DATASET")

var (
	bqClient = client.New(PEM_PATH)
)

func CheckMinBQCount(query string, minResults int) error {

	dataChan := make(chan client.Data)

	bqClient.PrintDebug = false
	go bqClient.AsyncQuery(100, DATASET, PROJECTID, query, dataChan)

	for {
		select {
		case d, ok := <-dataChan:
			if d.Err != nil {
				fmt.Println("Error with data: ", d.Err)
				return d.Err
			}
			if d.Rows != nil && d.Headers != nil {
				count := d.Rows[0][0].(string)
				results, err := strconv.Atoi(count)
				if err != nil {
					return fmt.Errorf("could not parse query result %s", err.Error())
				}
				if results < minResults {
					return fmt.Errorf("expected query to yield %d results but got only %d.", minResults, results)
				} else {
					return nil
				}
			}

			if !ok {
				return fmt.Errorf("bq watch: data channel closed.")
			}
		}
	}

}
