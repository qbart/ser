package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elbv2"
)

const DotRed = "\033[31;1m●\033[0m"
const DotGreen = "\033[32;1m●\033[0m"
const DotYellow = "\033[33;1m●\033[0m"

func awsFindTag(tags []*ec2.Tag, lookupValue string) *string {
	for _, tag := range tags {
		if *tag.Key == lookupValue {
			return tag.Value
		}
	}

	return nil
}

func awsCheckErrors(err error) bool {
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			fmt.Println(err.Error())
		}
		return true
	}

	return false
}

func dashify(str *string) string {
	if str == nil {
		return "-"
	}

	return *str
}

func awsInstanceStatusDot(state int64) string {
	// 0 (pending)
	// 16 (running)
	// 32 (shutting-down)
	// 48 (terminated)
	// 64 (stopping)
	// 80 (stopped)

	switch state {
	case 16:
		return DotGreen
	case 48, 80:
		return DotRed
	default:
		return DotYellow
	}
}

func awsLoadBalancerStatusDot(state string) string {
	// active | provisioning | active_impaired | failed

	switch state {
	case "active":
		return DotGreen
	case "failed":
		return DotRed
	default:
		return DotYellow
	}
}

func awsZonesToList(zones []*elbv2.AvailabilityZone) []string {
	result := make([]string, len(zones))

	for i, zone := range zones {
		result[i] = *zone.ZoneName
	}

	return result
}

func awsCopyList(list []*string) []string {
	result := make([]string, len(list))
	for i, str := range list {
		result[i] = *str
	}
	return result
}
