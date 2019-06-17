package main

import "time"

type AwsInstance struct {
	state       int64
	ami         string
	ipv4        string
	id          string
	kind        string
	zone        string
	name        string
	environment string
	launchTime  time.Time
}

type Dashboard struct {
	instances []*AwsInstance
}
