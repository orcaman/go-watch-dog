package models

type Server struct {
	Name          string
	PublicDnsName string
	DeployHash    string
	Tags          []string
}

type MonitoredService struct {
	Name    string
	Servers map[string]*Server
}
