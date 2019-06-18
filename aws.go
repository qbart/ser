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

	return result
}

func awsGetTargetGroups(sess *session.Session) []*AwsTargetGroup {
	client := elbv2.New(sess)
	input := &elbv2.DescribeTargetGroupsInput{}

	response, err := client.DescribeTargetGroups(input)
	awsCheckErrors(err)

	result := make([]*AwsTargetGroup, 0)

	for _, targetGroup := range response.TargetGroups {
		row := &AwsTargetGroup{
			arn:              *targetGroup.TargetGroupArn,
			name:             toS(targetGroup.TargetGroupName),
			port:             *targetGroup.Port,
			protocol:         toS(targetGroup.Protocol),
			kind:             toS(targetGroup.TargetType),
			loadBalancerArns: awsCopyList(targetGroup.LoadBalancerArns),
		}
		result = append(result, row)
	}

	return result
}

func awsGetTargetHealth(sess *session.Session, arn string) []*AwsTargetHealth {
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
