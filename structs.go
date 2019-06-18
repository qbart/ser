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

type AwsLoadBalancer struct {
	name   string
	dns    string
	state  string
	zones  []string
	kind   string
	scheme string
}

type Dashboard struct {
	instances     []*AwsInstance
	loadBalancers []*AwsLoadBalancer
}
