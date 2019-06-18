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
	arn    string
	name   string
	dns    string
	state  string
	zones  []string
	kind   string
	scheme string
}

type AwsTargetGroup struct {
	arn              string
	name             string
	port             int64
	protocol         string
	kind             string
	targets          []string
	loadBalancerArns []string
}

type Dashboard struct {
	instances     []*AwsInstance
	loadBalancers []*AwsLoadBalancer
	targetGroups  []*AwsTargetGroup
}
