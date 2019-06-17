package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func awsNewSession(region string, profile string) *session.Session {
	return session.Must(session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewSharedCredentials("", profile),
	}))
}

func awsGetInstances(sess *session.Session) []*AwsInstance {
	ec2Client := ec2.New(sess)
	input := &ec2.DescribeInstancesInput{}

	res, err := ec2Client.DescribeInstances(input)
	awsCheckErrors(err)

	result := make([]*AwsInstance, 0)

	for _, reservation := range res.Reservations {
		for _, instance := range reservation.Instances {
			awsInstance := &AwsInstance{
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
			result = append(result, awsInstance)
		}
	}

	return result
}
