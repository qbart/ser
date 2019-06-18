package main

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws/endpoints"
)

func main() {
	// box := tview.NewBox().SetBorder(true).SetTitle("Instances")

	// if err := tview.NewApplication().SetRoot(box, true).Run(); err != nil {
	// 	panic(err)
	// }

	session := awsNewSession(endpoints.EuWest1RegionID, "default")

	dashboard := &Dashboard{}
	dashboard.instances = awsGetInstances(session)
	dashboard.loadBalancers = awsGetLoadBalancers(session)
	dashboard.targetGroups = awsGetTargetGroups(session)

	for _, instance := range dashboard.instances {
		fmt.Println(awsInstanceStatusDot(instance.state))
		fmt.Println(instance.id)
		fmt.Println(instance.name)
		fmt.Println(instance.environment)
		fmt.Println(instance.ipv4)
		fmt.Println(instance.kind)
		fmt.Println(instance.ami)
		fmt.Println(instance.zone)
		fmt.Println(instance.launchTime)
		fmt.Println("----------")
	}

	for _, balancer := range dashboard.loadBalancers {
		fmt.Println(awsLoadBalancerStatusDot(balancer.state))
		fmt.Println(balancer.name)
		fmt.Println(balancer.dns)
		fmt.Println(balancer.kind)
		fmt.Println(balancer.scheme)
		fmt.Println(strings.Join(balancer.zones, ", "))
		fmt.Println("----------")
	}

	for _, target := range dashboard.targetGroups {
		fmt.Println(target.arn)
		fmt.Println(target.kind)
		fmt.Println(target.name)
		fmt.Println(target.port)
		fmt.Println(target.protocol)
		fmt.Println(strings.Join(target.loadBalancerArns, ", "))
		target.targets = awsGetTargetHealth(session, target.arn)

		for _, t := range target.targets {
			fmt.Printf("  %s\n", awsTargetHealthStatusDot(t.state))
			fmt.Printf("  %s\n", t.instanceId)
			fmt.Printf("  %d\n", t.port)
			fmt.Printf("  %s\n", t.reason)
			fmt.Printf("  %s\n", t.zone)
			fmt.Println("  --")
		}
		fmt.Println("----------")
	}
}
