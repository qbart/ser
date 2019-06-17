package main

import (
	"fmt"

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

	for _, instance := range dashboard.instances {
		fmt.Println(instance.id)
		fmt.Println(instance.name)
		fmt.Println(instance.environment)
		fmt.Println(instance.ipv4)
		fmt.Println(instance.kind)
		fmt.Println(awsInstanceStatusDot(instance.state))
		fmt.Println(instance.ami)
		fmt.Println(instance.zone)
		fmt.Println(instance.launchTime)
	}
}
