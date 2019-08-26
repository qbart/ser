package main

import (
	"log"

	"github.com/alecthomas/kingpin"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

func emptyTableData() [][]string {
	return [][]string{[]string{""}}
}

var (
	ColorOrange ui.Color = 202
	ColorPink   ui.Color = 198
	tabNames    []string = []string{"Pipelines", "Instances", "Target groups", "Load balancers"}
)

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

	messagesCh := make(chan string)
	codePipelinesCh := make(chan [][]string)
	instancesCh := make(chan [][]string)
	targetGroupsCh := make(chan [][]string)
	loadBalancersCh := make(chan [][]string)
	go awsPoolingLoop(*argProfile, messagesCh, codePipelinesCh, instancesCh, targetGroupsCh, loadBalancersCh)

	uiTables := make([]*widgets.Table, len(tabNames))
	uiGrids := make([]*ui.Grid, len(tabNames))

	termWidth, termHeight := ui.TerminalDimensions()

	uiFooter := widgets.NewParagraph()
	uiFooter.SetRect(0, termHeight-1, termWidth, termHeight)
	uiFooter.Border = false
	uiFooter.TextStyle.Fg = ColorPink

	for i := 0; i < len(tabNames); i++ {
		tab := widgets.NewTable()
		tab.Border = true
		tab.TextStyle = ui.NewStyle(ui.ColorWhite)
		tab.RowSeparator = false
		tab.FillRow = true
		tab.Title = tabNames[i]

		grid := ui.NewGrid()
		grid.Border = false
		grid.SetRect(0, 0, termWidth, termHeight-2)
		grid.Set(
			ui.NewRow(1.0, ui.NewCol(1.0, tab)),
		)

		uiGrids[i] = grid
		uiTables[i] = tab
		uiTables[i].Rows = emptyTableData()
	}

	for i := 0; i < len(tabNames); i++ {
		uiTables[i].RowStyles[0] = ui.NewStyle(ui.ColorWhite, ColorOrange, ui.ModifierBold)
	}

	uiTabs := widgets.NewTabPane(tabNames...)
	uiTabs.SetRect(0, 0, termWidth, termHeight-1)
	uiTabs.Border = false
	uiTabs.Block.Inner.Min.Y = termHeight - 2
	uiTabs.ActiveTabStyle.Fg = ColorOrange
	uiTabs.ActiveTabStyle.Modifier = ui.ModifierBold

	uiRenderTab := func(i int) {
		ui.Clear()
		ui.Render(uiTabs, uiGrids[i], uiFooter)
	}
	uiEvents := ui.PollEvents()

	for {
		select {
		case msg := <-messagesCh:
			uiFooter.Text = msg
			uiRenderTab(uiTabs.ActiveTabIndex)

		case rows := <-codePipelinesCh:
			uiTables[0].Rows = rows
			uiRenderTab(uiTabs.ActiveTabIndex)

		case rows := <-instancesCh:
			uiTables[1].Rows = rows
			uiRenderTab(uiTabs.ActiveTabIndex)

		case rows := <-targetGroupsCh:
			uiTables[2].Rows = rows
			uiRenderTab(uiTabs.ActiveTabIndex)

		case rows := <-loadBalancersCh:
			uiTables[3].Rows = rows
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
