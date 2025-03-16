package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"mytunnel/internal/config"
	"mytunnel/internal/ssh"
)

// UI represents the terminal user interface
type UI struct {
	app           *tview.Application
	table         *tview.Table
	statusBar     *tview.TextView
	tunnelManager *ssh.TunnelManager
	bastion       *config.BastionConfig
	ports         []int
	filter        string
	mainFlex      *tview.Flex // Add this field to store the main layout
}

// NewUI creates a new terminal UI
func NewUI(tunnelManager *ssh.TunnelManager, bastion *config.BastionConfig) *UI {
	ui := &UI{
		app:           tview.NewApplication(),
		tunnelManager: tunnelManager,
		bastion:       bastion,
		ports:         make([]int, 0),
	}

	ui.setupUI()
	return ui
}

// setupUI initializes the UI components
func (ui *UI) setupUI() {
	// Create main table
	ui.table = tview.NewTable().
		SetBorders(true).
		SetSelectable(true, false)

	// Create status bar
	ui.statusBar = tview.NewTextView().
		SetDynamicColors(true).
		SetText("Press '?' for help")

	// Create layout
	ui.mainFlex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(ui.table, 0, 1, true).
		AddItem(ui.statusBar, 1, 1, false)

	// Set up key bindings
	ui.app.SetInputCapture(ui.handleInput)

	// Set up table headers
	ui.table.SetCell(0, 0, tview.NewTableCell("Local Port").SetSelectable(false).SetTextColor(tcell.ColorYellow))
	ui.table.SetCell(0, 1, tview.NewTableCell("Remote Port").SetSelectable(false).SetTextColor(tcell.ColorYellow))
	ui.table.SetCell(0, 2, tview.NewTableCell("Status").SetSelectable(false).SetTextColor(tcell.ColorYellow))

	ui.app.SetRoot(ui.mainFlex, true)
}

// handleInput processes keyboard input
func (ui *UI) handleInput(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyEscape:
		ui.app.Stop()
		return nil
	case tcell.KeyRune:
		switch event.Rune() {
		case 'q':
			ui.app.Stop()
			return nil
		case 'j':
			row, _ := ui.table.GetSelection()
			if row < ui.table.GetRowCount()-1 {
				ui.table.Select(row+1, 0)
			}
			return nil
		case 'k':
			row, _ := ui.table.GetSelection()
			if row > 1 {
				ui.table.Select(row-1, 0)
			}
			return nil
		case '/':
			ui.showFilterPrompt()
			return nil
		case 't':
			ui.toggleTunnelView()
			return nil
		case 'd':
			ui.closeTunnel()
			return nil
		case '?':
			ui.showHelp()
			return nil
		case ' ', '\r':
			ui.openTunnel()
			return nil
		}
	}
	return event
}

// showFilterPrompt shows a prompt for filtering ports
func (ui *UI) showFilterPrompt() {
	form := tview.NewForm()
	form.AddInputField("Filter", "", 30, nil, func(text string) {
		ui.filter = text
		ui.updateTable()
		ui.app.SetRoot(ui.mainFlex, true)
	})
	form.AddButton("Cancel", func() {
		ui.app.SetRoot(ui.mainFlex, true)
	})
	form.SetBorder(true)
	form.SetTitle(" Filter Ports ")
	
	// Center the form
	flex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().
			AddItem(nil, 0, 1, false).
			AddItem(form, 40, 1, true).
			AddItem(nil, 0, 1, false), 3, 1, true).
		AddItem(nil, 0, 1, false)
		
	ui.app.SetRoot(flex, true)
}

// toggleTunnelView switches between available ports and active tunnels
func (ui *UI) toggleTunnelView() {
	// Implementation depends on how you want to display active tunnels
	tunnels := ui.tunnelManager.ListTunnels()
	ui.table.Clear()

	// Redraw headers
	ui.table.SetCell(0, 0, tview.NewTableCell("Local Port").SetSelectable(false).SetTextColor(tcell.ColorYellow))
	ui.table.SetCell(0, 1, tview.NewTableCell("Remote Port").SetSelectable(false).SetTextColor(tcell.ColorYellow))
	ui.table.SetCell(0, 2, tview.NewTableCell("Status").SetSelectable(false).SetTextColor(tcell.ColorYellow))

	for i, tunnel := range tunnels {
		ui.table.SetCell(i+1, 0, tview.NewTableCell(fmt.Sprintf("%d", tunnel.LocalPort)))
		ui.table.SetCell(i+1, 1, tview.NewTableCell(fmt.Sprintf("%d", tunnel.RemotePort)))
		ui.table.SetCell(i+1, 2, tview.NewTableCell("Active").SetTextColor(tcell.ColorGreen))
	}
}

// openTunnel opens a new SSH tunnel for the selected port
func (ui *UI) openTunnel() {
	row, _ := ui.table.GetSelection()
	if row <= 0 {
		return
	}

	localPortStr := ui.table.GetCell(row, 0).Text
	remotePortStr := ui.table.GetCell(row, 1).Text

	localPort, _ := strconv.Atoi(localPortStr)
	remotePort, _ := strconv.Atoi(remotePortStr)

	go func() {
		if err := ui.tunnelManager.CreateTunnel(localPort, remotePort, ui.bastion); err != nil {
			ui.showError(fmt.Sprintf("Failed to create tunnel: %v", err))
			return
		}
		ui.app.QueueUpdateDraw(func() {
			ui.table.GetCell(row, 2).SetText("Active").SetTextColor(tcell.ColorGreen)
		})
	}()
}

// closeTunnel closes the selected tunnel
func (ui *UI) closeTunnel() {
	row, _ := ui.table.GetSelection()
	if row <= 0 {
		return
	}

	localPortStr := ui.table.GetCell(row, 0).Text
	localPort, _ := strconv.Atoi(localPortStr)

	if err := ui.tunnelManager.CloseTunnel(localPort); err != nil {
		ui.showError(fmt.Sprintf("Failed to close tunnel: %v", err))
		return
	}

	ui.table.GetCell(row, 2).SetText("Closed").SetTextColor(tcell.ColorRed)
}

// showError displays an error message in the status bar
func (ui *UI) showError(msg string) {
	ui.app.QueueUpdateDraw(func() {
		ui.statusBar.SetText(fmt.Sprintf("[red]Error: %s[-]", msg))
	})
}

// showHelp displays the help dialog
func (ui *UI) showHelp() {
	text := `
[yellow]Keyboard Controls:[-]
j/k - Navigate up/down
Enter/Space - Open tunnel
t - Toggle tunnel view
d - Close tunnel
/ - Filter ports
q/Esc - Quit
? - Show this help
`
	modal := tview.NewModal().
		SetText(text).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			ui.app.SetRoot(ui.mainFlex, true)
		})

	ui.app.SetRoot(modal, true)
}

// updateTable updates the table with filtered ports
func (ui *UI) updateTable() {
	ui.table.Clear()

	// Set headers
	ui.table.SetCell(0, 0, tview.NewTableCell("Local Port").SetSelectable(false).SetTextColor(tcell.ColorYellow))
	ui.table.SetCell(0, 1, tview.NewTableCell("Remote Port").SetSelectable(false).SetTextColor(tcell.ColorYellow))
	ui.table.SetCell(0, 2, tview.NewTableCell("Status").SetSelectable(false).SetTextColor(tcell.ColorYellow))

	row := 1
	for _, port := range ui.ports {
		if ui.filter != "" && !strings.Contains(fmt.Sprintf("%d", port), ui.filter) {
			continue
		}

		ui.table.SetCell(row, 0, tview.NewTableCell(fmt.Sprintf("%d", port)))
		ui.table.SetCell(row, 1, tview.NewTableCell(fmt.Sprintf("%d", port)))
		ui.table.SetCell(row, 2, tview.NewTableCell("Available").SetTextColor(tcell.ColorWhite))
		row++
	}
}

// Run starts the UI
func (ui *UI) Run() error {
	return ui.app.Run()
}

// Stop stops the UI
func (ui *UI) Stop() {
	ui.app.Stop()
}

// SetPorts updates the available ports list
func (ui *UI) SetPorts(ports []int) {
	ui.ports = ports
	ui.updateTable()
} 