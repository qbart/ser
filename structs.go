package main

import "time"

type AwsCodePipeline struct {
	name   string
	stages []*AwsCodePipelineStage
}

type AwsCodePipelineStage struct {
	name    string
	actions []*AwsCodePipelineAction
}

type AwsCodePipelineAction struct {
	name   string
	status string
}

type AwsInstance struct {
	state       int64
	ami         string
	ipv4        string
	ipv4private string
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
	loadBalancerArns []string
	targets          []*AwsTargetHealth
}

type AwsTargetHealth struct {
	instanceId string
	state      string
	zone       string
	port       int64
	reason     string
}

type Dashboard struct {
	pipelines      []*AwsCodePipeline
	instances      []*AwsInstance
	loadBalancers  []*AwsLoadBalancer
	targetGroups   []*AwsTargetGroup
	zoneByInstance map[string]string
}
