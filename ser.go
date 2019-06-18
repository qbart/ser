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
}
