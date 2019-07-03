package main

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go/aws/endpoints"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

func main() {
	profile := "default"
	dashboard := &Dashboard{}
	session := awsNewSession(endpoints.EuWest1RegionID, profile)
	dashboard.zoneByInstance = make(map[string]string)
	dashboard.instances = awsGetInstances(session)
	dashboard.loadBalancers = awsGetLoadBalancers(session)
	dashboard.targetGroups = awsGetTargetGroups(session)
	for _, target := range dashboard.targetGroups {
		target.targets = awsGetTargetHealth(session, target.arn)
	}

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

	if err := ui.Init(); err != nil {
		log.Fatalf("Failed to initialize ser: %v", err)
	}
	defer ui.Close()

	instancesData := [][]string{
		[]string{"Environment", "State", "Name", "Type", "IPv4", "Zone", "ID", "AMI", "Launch time"},
	}
	for _, instance := range dashboard.instances {
		row := []string{
			instance.environment,
			awsInstanceStatus(instance.state),
			instance.name,
			instance.kind,
			instance.ipv4,
			instance.zone,
			instance.id,
			instance.ami,
			instance.launchTime.Format("02-01-2006 15:04 MST"),
		}
		dashboard.zoneByInstance[instance.id] = instance.zone
		instancesData = append(instancesData, row)
	}

	tgData := [][]string{
		[]string{"Instance ID", "Zone", "Port"},
	}

	for _, tg := range dashboard.targetGroups {
		row := []string{
			fmt.Sprintf("%s (%s)", tg.name, tg.kind),
			"",
			fmt.Sprintf("%s -> %d", tg.protocol, tg.port),
		}
		tgData = append(tgData, row)

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
			tgData = append(tgData, row)
		}
	}

	lbData := [][]string{
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

		lbData = append(lbData, row)
	}

	table1 := widgets.NewTable()
	table1.Border = false
	table1.Rows = instancesData
	table1.TextStyle = ui.NewStyle(ui.ColorWhite)
	table1.RowSeparator = false
	table1.FillRow = true
	table1.RowStyles[0] = ui.NewStyle(ui.ColorWhite, ui.ColorBlue, ui.ModifierBold)

	table2 := widgets.NewTable()
	table2.Border = false
	table2.Rows = tgData
	table2.TextStyle = ui.NewStyle(ui.ColorWhite)
	table2.RowSeparator = false
	table2.FillRow = true
	table2.RowStyles[0] = ui.NewStyle(ui.ColorWhite, ui.ColorBlue, ui.ModifierBold)

	table3 := widgets.NewTable()
	table3.Border = false
	table3.Rows = lbData
	table3.TextStyle = ui.NewStyle(ui.ColorWhite)
	table3.RowSeparator = false
	table3.FillRow = true
	table3.RowStyles[0] = ui.NewStyle(ui.ColorWhite, ui.ColorBlue, ui.ModifierBold)

	termWidth, termHeight := ui.TerminalDimensions()
	tabpane := widgets.NewTabPane("Instances", "Target groups", "Load balancers")
	tabpane.SetRect(0, 0, termWidth, 1)
	tabpane.Border = false
	grid := ui.NewGrid()
	grid.Border = false
	grid.SetRect(0, 1, termWidth, termHeight-1)

	renderTab := func() {
		ui.Clear()
		switch tabpane.ActiveTabIndex {
		case 0:
			grid.Set(
				ui.NewRow(1.0, ui.NewCol(1.0, table1)),
			)

			ui.Render(tabpane, grid)
			ui.Render(table1)

		case 1:
			grid.Set(
				ui.NewRow(1.0, ui.NewCol(1.0, table2)),
			)

			ui.Render(tabpane, grid)
			ui.Render(table2)

		case 2:
			grid.Set(
				ui.NewRow(1.0, ui.NewCol(1.0, table3)),
			)

			ui.Render(tabpane, grid)
			ui.Render(table3)
		}
	}

	renderTab()
	uiEvents := ui.PollEvents()

	for {
		e := <-uiEvents
		switch e.ID {
		case "q", "<C-c>":
			return
		case "j":
			tabpane.FocusLeft()
			renderTab()
		case ";":
			tabpane.FocusRight()
			renderTab()
		}
	}
}
