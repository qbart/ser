package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

func main() {
	dashboard := &Dashboard{}
	session := awsNewSession(endpoints.EuWest1RegionID, "default")
	dashboard.instances = awsGetInstances(session)
	sort.Slice(dashboard.instances, func(i, j int) bool {
		a := dashboard.instances[i]
		b := dashboard.instances[j]
		if a.environment == b.environment {
			if a.name == b.name {
				return a.zone < b.zone
			}
			return a.name < b.name
		}
		return a.environment > b.environment
	})

	app := tview.NewApplication()
	pages := tview.NewPages()

	instancesTable := tview.NewTable().SetBorders(true)

	pages.AddPage("Instances", instancesTable, false, true)
	instancesTable.SetCell(0, 0, tview.NewTableCell("Environment").SetTextColor(tcell.ColorYellow))
	instancesTable.SetCell(0, 1, tview.NewTableCell("State").SetTextColor(tcell.ColorYellow))
	instancesTable.SetCell(0, 2, tview.NewTableCell("Name").SetTextColor(tcell.ColorYellow))
	instancesTable.SetCell(0, 3, tview.NewTableCell("Type").SetTextColor(tcell.ColorYellow))
	instancesTable.SetCell(0, 4, tview.NewTableCell("IPv4").SetTextColor(tcell.ColorYellow))
	instancesTable.SetCell(0, 5, tview.NewTableCell("Zone").SetTextColor(tcell.ColorYellow))
	instancesTable.SetCell(0, 6, tview.NewTableCell("ID").SetTextColor(tcell.ColorYellow))
	instancesTable.SetCell(0, 7, tview.NewTableCell("AMI").SetTextColor(tcell.ColorYellow))
	instancesTable.SetCell(0, 8, tview.NewTableCell("Launch time").SetTextColor(tcell.ColorYellow))

	for i, instance := range dashboard.instances {
		instancesTable.SetCell(i+1, 0, tview.NewTableCell(instance.environment))
		instancesTable.SetCell(i+1, 1, tview.NewTableCell(Dot).SetTextColor(awsInstanceStatusColor(instance.state)).SetAlign(tview.AlignCenter))
		instancesTable.SetCell(i+1, 2, tview.NewTableCell(instance.name))
		instancesTable.SetCell(i+1, 3, tview.NewTableCell(instance.kind))
		instancesTable.SetCell(i+1, 4, tview.NewTableCell(instance.ipv4))
		instancesTable.SetCell(i+1, 5, tview.NewTableCell(instance.zone))
		instancesTable.SetCell(i+1, 6, tview.NewTableCell(instance.id))
		instancesTable.SetCell(i+1, 7, tview.NewTableCell(instance.ami))
		instancesTable.SetCell(i+1, 8, tview.NewTableCell(instance.launchTime.Format("02-01-2006 15:04 MST")))
	}

	flex := tview.NewFlex().AddItem(instancesTable, 0, 1, false)

	if err := app.SetRoot(flex, true).SetFocus(pages).Run(); err != nil {
		panic(err)
	}

	// session := awsNewSession(endpoints.EuWest1RegionID, "default")
	// dashboard.instances = awsGetInstances(session)
	// dashboard.loadBalancers = awsGetLoadBalancers(session)
	// dashboard.targetGroups = awsGetTargetGroups(session)

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
		// target.targets = awsGetTargetHealth(session, target.arn)

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
