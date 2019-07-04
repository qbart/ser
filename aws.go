package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elbv2"
)

func awsPoolingLoop(
	profile string,
	// out
	instances chan [][]string,
	targetGroups chan [][]string,
	loadBalancers chan [][]string,
) {
	session := awsNewSession(endpoints.EuWest1RegionID, profile)

	dashboard := &Dashboard{}
	dashboard.zoneByInstance = make(map[string]string)

	every15s := time.Tick(15 * time.Second)
	every30s := time.Tick(30 * time.Second)
	every60s := time.Tick(60 * time.Second)

	instancesAPI := make(chan []*AwsInstance)
	targetGroupsAPI := make(chan []*AwsTargetGroup)
	loadBalancersAPI := make(chan []*AwsLoadBalancer)

	go awsGetInstances(session, instancesAPI)
	go awsGetTargetGroups(session, targetGroupsAPI)
	go awsGetLoadBalancers(session, loadBalancersAPI)

	for {
		select {
		case <-every15s:
			go awsGetTargetGroups(session, targetGroupsAPI)

		case <-every30s:
			go awsGetInstances(session, instancesAPI)

		case <-every60s:
			go awsGetLoadBalancers(session, loadBalancersAPI)

		case dashboard.instances = <-instancesAPI:
			dashboard.zoneByInstance = make(map[string]string)
			rows := [][]string{
				[]string{"Environment", "State", "Name", "Type", "IPv4", "Zone", "ID", "AMI", "Launch time"},
			}

			for _, instance := range dashboard.instances {
				row := []string{
					instance.environment,
					awsInstanceStatus(instance.state),
					instance.name,
					instance.kind,
					instance.ipv4,
					instance.zone,
					instance.id,
					instance.ami,
					instance.launchTime.Format("02-01-2006 15:04 MST"),
				}
				rows = append(rows, row)
				dashboard.zoneByInstance[instance.id] = instance.zone
			}

			instances <- rows

		case dashboard.targetGroups = <-targetGroupsAPI:
			rows := [][]string{
				[]string{"Instance ID", "Zone", "Port"},
			}

			for _, tg := range dashboard.targetGroups {
				row := []string{
					fmt.Sprintf("%s (%s)", tg.name, tg.kind),
					"",
					fmt.Sprintf("%s -> %d", tg.protocol, tg.port),
				}
				rows = append(rows, row)

				for _, t := range tg.targets {
					reason := fmt.Sprintf(" (%s)", t.reason)
					if len(t.reason) == 0 {
						reason = ""
					}
					zone := t.zone
					if len(zone) == 0 {
						zone = dashboard.zoneByInstance[t.instanceId]
					}
					row := []string{
						fmt.Sprintf("  %s", t.instanceId),
						zone,
						fmt.Sprintf("%6d %s%s", t.port, t.state, reason),
					}
					rows = append(rows, row)
				}
			}

			targetGroups <- rows

		case dashboard.loadBalancers = <-loadBalancersAPI:
			rows := [][]string{
				[]string{"State", "Name", "Dns", "Kind", "Scheme", "Zones"},
			}

			for _, balancer := range dashboard.loadBalancers {
				row := []string{
					balancer.state,
					balancer.name,
					balancer.dns,
					balancer.kind,
					balancer.scheme,
					strings.Join(balancer.zones, ", "),
				}
				rows = append(rows, row)
			}

			loadBalancers <- rows
		}
	}
}

func awsNewSession(region string, profile string) *session.Session {
	return session.Must(session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewSharedCredentials("", profile),
	}))
}

func awsGetInstances(sess *session.Session, out chan<- []*AwsInstance) {
	client := ec2.New(sess)
	input := &ec2.DescribeInstancesInput{}

	response, err := client.DescribeInstances(input)
	awsCheckErrors(err)

	result := make([]*AwsInstance, 0)

	for _, reservation := range response.Reservations {
		for _, instance := range reservation.Instances {
			row := &AwsInstance{
				id:          toS(instance.InstanceId),
				ipv4:        toS(instance.PublicIpAddress),
				kind:        toS(instance.InstanceType),
				state:       *instance.State.Code,
				ami:         toS(instance.ImageId),
				zone:        toS(instance.Placement.AvailabilityZone),
				launchTime:  *instance.LaunchTime,
				name:        toS(awsFindTag(instance.Tags, "Name")),
				environment: toS(awsFindTag(instance.Tags, "Environment")),
			}
			result = append(result, row)
		}
	}
	awsSortInstances(result)

	out <- result
}

func awsGetLoadBalancers(sess *session.Session, out chan<- []*AwsLoadBalancer) {
	client := elbv2.New(sess)
	input := &elbv2.DescribeLoadBalancersInput{}

	response, err := client.DescribeLoadBalancers(input)
	awsCheckErrors(err)

	result := make([]*AwsLoadBalancer, 0)

	for _, loadBalancer := range response.LoadBalancers {
		row := &AwsLoadBalancer{
			arn:    *loadBalancer.LoadBalancerArn,
			name:   toS(loadBalancer.LoadBalancerName),
			dns:    toS(loadBalancer.DNSName),
			kind:   toS(loadBalancer.Type),
			scheme: toS(loadBalancer.Scheme),
			state:  toS(loadBalancer.State.Code),
			zones:  awsZonesToList(loadBalancer.AvailabilityZones),
		}
		result = append(result, row)
	}

	out <- result
}

func awsGetTargetGroups(sess *session.Session, out chan<- []*AwsTargetGroup) {
	client := elbv2.New(sess)
	input := &elbv2.DescribeTargetGroupsInput{}

	type Row struct {
		i       int
		targets []*AwsTargetHealth
	}

	response, err := client.DescribeTargetGroups(input)
	awsCheckErrors(err)

	result := make([]*AwsTargetGroup, 0)
	targets := make(chan Row, len(response.TargetGroups))

	for i, targetGroup := range response.TargetGroups {
		row := &AwsTargetGroup{
			arn:              *targetGroup.TargetGroupArn,
			name:             toS(targetGroup.TargetGroupName),
			port:             *targetGroup.Port,
			protocol:         toS(targetGroup.Protocol),
			kind:             toS(targetGroup.TargetType),
			loadBalancerArns: awsCopyList(targetGroup.LoadBalancerArns),
		}
		result = append(result, row)

		go func(arn string, i int, out chan Row) {
			rows := awsGetTargetHealths(sess, arn)
			out <- Row{
				i:       i,
				targets: rows,
			}
		}(row.arn, i, targets)
	}

	for i := 0; i < len(response.TargetGroups); i++ {
		row := <-targets
		result[row.i].targets = row.targets
	}

	out <- result
}

func awsGetTargetHealths(sess *session.Session, arn string) []*AwsTargetHealth {
	client := elbv2.New(sess)
	input := &elbv2.DescribeTargetHealthInput{
		TargetGroupArn: aws.String(arn),
	}

	response, err := client.DescribeTargetHealth(input)
	awsCheckErrors(err)

	result := make([]*AwsTargetHealth, 0)

	for _, health := range response.TargetHealthDescriptions {
		row := &AwsTargetHealth{
			instanceId: toS(health.Target.Id),
			state:      toS(health.TargetHealth.State),
			port:       *health.Target.Port,
			reason:     toS(health.TargetHealth.Description),
			zone:       toS(health.Target.AvailabilityZone),
		}
		result = append(result, row)
	}

	return result
}
