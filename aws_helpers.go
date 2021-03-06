package main

import (
	"fmt"
	"sort"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elbv2"
)

func awsFindTag(tags []*ec2.Tag, lookupValue string) *string {
	for _, tag := range tags {
		if *tag.Key == lookupValue {
			return tag.Value
		}
	}

	return nil
}

func awsCheckErrors(out chan string, err error) {
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				out <- aerr.Error()
			}
		} else {
			out <- err.Error()
		}
	}
}

func toS(str *string) string {
	if str == nil {
		return ""
	}

	return *str
}

func awsInstanceStatus(state int64) string {
	switch state {
	case 16:
		return "running"
	case 32:
		return "shutting-down"
	case 48:
		return "terminated"
	case 64:
		return "stopping"
	case 80:
		return "stopped"
	default: // 0
		return "pending"
	}
}

func awsPipelineActionsToList(actions []*AwsCodePipelineAction) []string {
	result := make([]string, len(actions))

	for i, action := range actions {
		result[i] = fmt.Sprintf("%s (%s)", action.name, action.status)
	}

	return result
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

func awsSortInstances(list []*AwsInstance) {
	sort.Slice(list, func(i, j int) bool {
		a := list[i]
		b := list[j]
		if a.environment == b.environment {
			if a.name == b.name {
				return a.zone < b.zone
			}
			return a.name < b.name
		}
		return a.environment > b.environment
	})
}
