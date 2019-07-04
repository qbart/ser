package main

import (
	"log"

	"github.com/alecthomas/kingpin"
	"github.com/gizak/termui/v3"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

func emptyTableData() [][]string {
	return [][]string{[]string{""}}
}

var (
	argProfile = kingpin.
		Arg("profile", "AWS profile name").
		Default("default").
		String()
)

func main() {
	if err := ui.Init(); err != nil {
		log.Fatalf("Failed to initialize ser: %v", err)
	}
	defer ui.Close()

	kingpin.Parse()

	instancesCh := make(chan [][]string)
	targetGroupsCh := make(chan [][]string)
	loadBalancersCh := make(chan [][]string)
	go awsPoolingLoop(*argProfile, instancesCh, targetGroupsCh, loadBalancersCh)

	uiTables := make([]*widgets.Table, 3)
	uiGrids := make([]*termui.Grid, 3)

	termWidth, termHeight := ui.TerminalDimensions()

	for i := 0; i < len(uiTables); i++ {
		tab := widgets.NewTable()
		tab.Border = false
		tab.TextStyle = ui.NewStyle(ui.ColorWhite)
		tab.RowSeparator = false
		tab.FillRow = true
		tab.Rows = [][]string{}

		grid := ui.NewGrid()
		grid.Border = false
		grid.SetRect(0, 1, termWidth, termHeight-1)
		grid.Set(
			ui.NewRow(1.0, ui.NewCol(1.0, tab)),
		)

		uiGrids[i] = grid
		uiTables[i] = tab
	}
	uiTables[0].Rows = emptyTableData()
	uiTables[1].Rows = emptyTableData()
	uiTables[2].Rows = emptyTableData()

	for i := 0; i < len(uiTables); i++ {
		uiTables[i].RowStyles[0] = ui.NewStyle(ui.ColorWhite, ui.ColorBlue, ui.ModifierBold)
	}

	uiTabs := widgets.NewTabPane("Instances", "Target groups", "Load balancers")
	uiTabs.SetRect(0, 0, termWidth, 1)
	uiTabs.Border = false

	uiRenderTab := func(i int) {
		ui.Clear()
		ui.Render(uiTabs, uiGrids[i])
	}
	uiEvents := ui.PollEvents()

	for {
		select {
		case rows := <-instancesCh:
			uiTables[0].Rows = rows
			uiRenderTab(uiTabs.ActiveTabIndex)

		case rows := <-targetGroupsCh:
			uiTables[1].Rows = rows
			uiRenderTab(uiTabs.ActiveTabIndex)

		case rows := <-loadBalancersCh:
			uiTables[2].Rows = rows
			uiRenderTab(uiTabs.ActiveTabIndex)

		case e := <-uiEvents:
			switch e.ID {
			case "q", "<C-c>":
				return
			case "j", "<Left>":
				uiTabs.FocusLeft()
				uiRenderTab(uiTabs.ActiveTabIndex)
			case ";", "<Right>":
				uiTabs.FocusRight()
				uiRenderTab(uiTabs.ActiveTabIndex)
			}
		}
	}
}
