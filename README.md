# go-watch-dog

Get notifications according to configurable conditions.


## Default configuration
Login to dashboard with basic auth

Username: admin
Password: admin 
(You should [change](http://www.htaccesstools.com/articles/htpasswd/) it in the .htpasswd file before deployment to prod)

1. You have to use your [service account](https://developers.google.com/identity/protocols/OAuth2ServiceAccount#overview) json 

2. You have to set environment variable for bigquery/gcs : 
```
set WATCHDOG_BQPROJECT=<my google project id>
set WATCHDOG_DATASET=<relevant dataset>
```

## supported test types

http checks:
- latency
- statuscode
- change 

deploy checks:
- deployhash (check that under the given instance group all hashes are the same)
- gcedeployhash (same as above but for GCE)
- minInstanceCount (check that under the given instance group there are at least the given number of instances)

bigquery:
- bqCount (runs a query in bq and checks if count is greater than result)

## deployhash, minInstanceCount
These type of checks require credentials to AWS. These are normally stored at ~/.aws/credentials.

## TODO

support the following test types:
- gcsMaxFiles
- commitSha

Add stuff:
- Connect to DB, store every result to generate uptime stats
- Expose API to get status per check 

## Example configuration

see watchdog.config file


## License

MIT (see [LICENSE](https://github.com/streamrail/go-watch-dog/blob/master/LICENSE.txt) file)
