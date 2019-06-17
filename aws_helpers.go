package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
)

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
		return "\033[32;1m●\033[0m"
	case 48, 80:
		return "\033[31;1m●\033[0m"
	default:
		return "\033[33;1m●\033[0m"
	}
}
