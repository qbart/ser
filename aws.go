package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/codepipeline"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elbv2"
)

func awsPoolingLoop(
	profile string,
	region string,
	// out
	messages chan string,
	codePipelines chan [][]string,
	instances chan [][]string,
	targetGroups chan [][]string,
	loadBalancers chan [][]string,
) {
	session := awsNewSession(region, profile)

	dashboard := &Dashboard{}
	dashboard.zoneByInstance = make(map[string]string)

	every15s := time.Tick(15 * time.Second)
	every30s := time.Tick(30 * time.Second)
	every60s := time.Tick(60 * time.Second)

	codePipelinesAPI := make(chan []*AwsCodePipeline)
	instancesAPI := make(chan []*AwsInstance)
	targetGroupsAPI := make(chan []*AwsTargetGroup)
	loadBalancersAPI := make(chan []*AwsLoadBalancer)

	go awsGetCodePipelines(session, messages, codePipelinesAPI)
	go awsGetInstances(session, messages, instancesAPI)
	go awsGetTargetGroups(session, messages, targetGroupsAPI)
	go awsGetLoadBalancers(session, messages, loadBalancersAPI)

	for {
		select {
		case <-every15s:
			go awsGetCodePipelines(session, messages, codePipelinesAPI)
			go awsGetTargetGroups(session, messages, targetGroupsAPI)

		case <-every30s:
			go awsGetInstances(session, messages, instancesAPI)

		case <-every60s:
			go awsGetLoadBalancers(session, messages, loadBalancersAPI)

		case dashboard.pipelines = <-codePipelinesAPI:
			rows := [][]string{
				[]string{"Name", "Stage", ""},
			}
			for _, pipeline := range dashboard.pipelines {
				row := []string{
					pipeline.name,
				}
				rows = append(rows, row)

				for _, stage := range pipeline.stages {
					row := []string{
						"",
						stage.name,
						strings.Join(awsPipelineActionsToList(stage.actions), " | "),
					}
					rows = append(rows, row)
				}
			}

			codePipelines <- rows

		case dashboard.instances = <-instancesAPI:
			dashboard.zoneByInstance = make(map[string]string)
			rows := [][]string{
				// []string{"Environment", "State", "Name", "Type", "IPv4", "IPv4 Priv", "Zone", "ID", "AMI", "Launch time"},
				{"ID", "State", "Type", "IP", "AMI", "Launch time"},
			}

			for _, instance := range dashboard.instances {
				row := []string{
					instance.id,
					awsInstanceStatus(instance.state),
					instance.kind,
					instance.ipv4private,
					instance.ami,
					instance.launchTime.Format("02-01-2006"),
				}
				rows = append(rows, row)

				publicIp := ""
				if instance.ipv4 != "" {
					publicIp = fmt.Sprint(instance.ipv4, " 🌍")
				}

				row = []string{
					instance.name,
					"",
					instance.zone,
					publicIp,
					"",
					instance.launchTime.Format(" 15:04 MST"),
				}
				rows = append(rows, row)

				row = []string{
					instance.environment,
					"",
					"",
					"",
					"",
					"",
				}
				rows = append(rows, row)
				dashboard.zoneByInstance[instance.id] = instance.zone
			}

			instances <- rows

		case dashboard.targetGroups = <-targetGroupsAPI:
			rows := [][]string{
				{"Instance ID", "Zone", "Port"},
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

func awsGetCodePipelines(sess *session.Session, msg chan string, out chan<- []*AwsCodePipeline) {
	client := codepipeline.New(sess)
	input := &codepipeline.ListPipelinesInput{}

	type Row struct {
		i      int
		stages []*AwsCodePipelineStage
	}

	response, err := client.ListPipelines(input)
	awsCheckErrors(msg, err)

	result := make([]*AwsCodePipeline, 0)
	actions := make(chan Row, len(response.Pipelines))

	for i, pipeline := range response.Pipelines {
		row := &AwsCodePipeline{
			name: toS(pipeline.Name),
		}
		result = append(result, row)

		go func(name string, i int, out chan Row) {
			rows := awsGetCodePipelineStages(sess, msg, name)
			out <- Row{
				i:      i,
				stages: rows,
			}
		}(row.name, i, actions)
	}

	for i := 0; i < len(response.Pipelines); i++ {
		row := <-actions
		result[row.i].stages = row.stages
	}

	out <- result
}

func awsGetInstances(sess *session.Session, msg chan string, out chan<- []*AwsInstance) {
	client := ec2.New(sess)
	input := &ec2.DescribeInstancesInput{}

	response, err := client.DescribeInstances(input)
	awsCheckErrors(msg, err)

	result := make([]*AwsInstance, 0)

	for _, reservation := range response.Reservations {
		for _, instance := range reservation.Instances {
			row := &AwsInstance{
				id:          toS(instance.InstanceId),
				ipv4:        toS(instance.PublicIpAddress),
				ipv4private: toS(instance.PrivateIpAddress),
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

func awsGetLoadBalancers(sess *session.Session, msg chan string, out chan<- []*AwsLoadBalancer) {
	client := elbv2.New(sess)
	input := &elbv2.DescribeLoadBalancersInput{}

	response, err := client.DescribeLoadBalancers(input)
	awsCheckErrors(msg, err)

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

func awsGetTargetGroups(sess *session.Session, msg chan string, out chan<- []*AwsTargetGroup) {
	client := elbv2.New(sess)
	input := &elbv2.DescribeTargetGroupsInput{}

	type Row struct {
		i       int
		targets []*AwsTargetHealth
	}

	response, err := client.DescribeTargetGroups(input)
	awsCheckErrors(msg, err)

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
			rows := awsGetTargetHealths(sess, msg, arn)
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

func awsGetCodePipelineStages(sess *session.Session, msg chan string, name string) []*AwsCodePipelineStage {
	client := codepipeline.New(sess)

	input := &codepipeline.GetPipelineInput{
		Name: aws.String(name),
	}
	response, err := client.GetPipeline(input)
	awsCheckErrors(msg, err)

	executionID := awsGetCodePipelineGetLastExecution(sess, msg, name)
	inputA := &codepipeline.ListActionExecutionsInput{
		PipelineName: aws.String(name),
		Filter: &codepipeline.ActionExecutionFilter{
			PipelineExecutionId: aws.String(executionID),
		},
	}

	responseA, errA := client.ListActionExecutions(inputA)
	awsCheckErrors(msg, errA)

	statuses := make(map[string]string)
	for _, details := range responseA.ActionExecutionDetails {
		statuses[toS(details.ActionName)] = toS(details.Status)
	}

	result := make([]*AwsCodePipelineStage, 0)

	for _, stage := range response.Pipeline.Stages {
		row := &AwsCodePipelineStage{
			name: toS(stage.Name),
		}
		row.actions = make([]*AwsCodePipelineAction, 0, len(stage.Actions))
		for _, action := range stage.Actions {
			status := "Didn't Run yet"
			if val, ok := statuses[toS(action.Name)]; ok {
				status = val
			}
			row.actions = append(row.actions, &AwsCodePipelineAction{
				name:   toS(action.Name),
				status: status,
			})
		}
		result = append(result, row)
	}

	return result
}

func awsGetCodePipelineGetLastExecution(sess *session.Session, msg chan string, name string) string {
	client := codepipeline.New(sess)

	input := &codepipeline.ListPipelineExecutionsInput{
		PipelineName: aws.String(name),
		MaxResults:   aws.Int64(1),
	}

	response, err := client.ListPipelineExecutions(input)
	awsCheckErrors(msg, err)

	for _, exec := range response.PipelineExecutionSummaries {
		return toS(exec.PipelineExecutionId)
	}

	return ""
}

func awsGetTargetHealths(sess *session.Session, msg chan string, arn string) []*AwsTargetHealth {
	client := elbv2.New(sess)
	input := &elbv2.DescribeTargetHealthInput{
		TargetGroupArn: aws.String(arn),
	}

	response, err := client.DescribeTargetHealth(input)
	awsCheckErrors(msg, err)

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
