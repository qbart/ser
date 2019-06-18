package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elbv2"
)

func awsNewSession(region string, profile string) *session.Session {
	return session.Must(session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewSharedCredentials("", profile),
	}))
}

func awsGetInstances(sess *session.Session) []*AwsInstance {
	client := ec2.New(sess)
	input := &ec2.DescribeInstancesInput{}

	response, err := client.DescribeInstances(input)
	awsCheckErrors(err)

	result := make([]*AwsInstance, 0)

	for _, reservation := range response.Reservations {
		for _, instance := range reservation.Instances {
			row := &AwsInstance{
				id:          dashify(instance.InstanceId),
				ipv4:        dashify(instance.PublicIpAddress),
				kind:        dashify(instance.InstanceType),
				state:       *instance.State.Code,
				ami:         dashify(instance.ImageId),
				zone:        dashify(instance.Placement.AvailabilityZone),
				launchTime:  *instance.LaunchTime,
				name:        dashify(awsFindTag(instance.Tags, "Name")),
				environment: dashify(awsFindTag(instance.Tags, "Environment")),
			}
			result = append(result, row)
		}
	}

	return result
}

func awsGetLoadBalancers(sess *session.Session) []*AwsLoadBalancer {
	client := elbv2.New(sess)
	input := &elbv2.DescribeLoadBalancersInput{}

	response, err := client.DescribeLoadBalancers(input)
	awsCheckErrors(err)

	result := make([]*AwsLoadBalancer, 0)

	for _, loadBalancer := range response.LoadBalancers {
		row := &AwsLoadBalancer{
			name:   dashify(loadBalancer.LoadBalancerName),
			dns:    dashify(loadBalancer.DNSName),
			kind:   dashify(loadBalancer.Type),
			scheme: dashify(loadBalancer.Scheme),
			state:  dashify(loadBalancer.State.Code),
			zones:  awsZonesToList(loadBalancer.AvailabilityZones),
		}
		result = append(result, row)
	}

	return result
}
