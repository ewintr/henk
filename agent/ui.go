package agent

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type MessageType string

const (
	TypeGeneral MessageType = "general"
	TypeHenk    MessageType = "henk"
	TypeUser    MessageType = "user"
	TypePrompt  MessageType = "prompt"
	TypeTool    MessageType = "tool"
	TypeError   MessageType = "error"
	TypeDebug   MessageType = "debug"
	TypeExit    MessageType = "exit"
)

type Message struct {
	Type MessageType
	Body string
}

type UI struct {
	app              *tview.Application
	userInputPane    *tview.TextArea
	conversationPane *tview.TextView
	statusPane       *tview.TextView
	conversation     []Message
	in               chan Message
	out              chan string
	cancel           context.CancelFunc
}

func NewUI(cancel context.CancelFunc) *UI {
	ui := &UI{
		app:    tview.NewApplication(),
		in:     make(chan Message),
		out:    make(chan string),
		cancel: cancel,
	}
	go ui.Run()

	return ui
}

func (ui *UI) In() chan Message { return ui.in }
func (ui *UI) Out() chan string { return ui.out }

func (ui *UI) Run() {
	ui.buildLayout()
	go ui.processInput()

	if err := ui.app.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}

func (ui *UI) processInput() {
	for msg := range ui.in {
		switch msg.Type {
		case TypeGeneral:
			ui.statusPane.SetText(msg.Body)
		case TypeHenk:
			ui.conversation = append(ui.conversation, msg)
			ui.redrawConversation()
		case TypeTool:
			ui.statusPane.SetText(fmt.Sprintf("tool: %s", msg.Body))
		case TypeError:
			ui.statusPane.SetText(fmt.Sprintf("error: %s", msg.Body))
		case TypeDebug:
			ui.statusPane.SetText(fmt.Sprintf("debug: %s", msg.Body))
		case TypeExit:
			ui.Close()
		}
		ui.app.Draw()
	}

}

func (ui *UI) Close() {
	ui.app.Stop()
	ui.cancel()
	close(ui.in)
	close(ui.out)
}

func (ui *UI) buildLayout() {
	ui.userInputPane = tview.NewTextArea()
	ui.userInputPane.SetBorder(true)

	ui.conversationPane = tview.NewTextView().
		SetDynamicColors(true)
	ui.statusPane = tview.NewTextView().SetText("status")
	layout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(ui.conversationPane, 0, 1, false).
		AddItem(ui.userInputPane, 10, 0, true).
		AddItem(ui.statusPane, 1, 0, false)

	ui.app.SetRoot(layout, true).
		SetInputCapture(ui.captureInput).
		EnableMouse(true)
}

func (ui *UI) captureInput(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyEnter:
		input := ui.userInputPane.GetText()
		if input == "" {
			return nil
		}
		msg := Message{
			Type: TypeUser,
			Body: input,
		}
		ui.conversation = append(ui.conversation, msg)
		ui.userInputPane.SetText("", true)
		ui.out <- input
		return nil
	case tcell.KeyPgUp:
		ui.app.SetFocus(ui.conversationPane)
		return tcell.NewEventKey(tcell.KeyCtrlB, 'b', tcell.ModCtrl)
	case tcell.KeyPgDn:
		ui.app.SetFocus(ui.conversationPane)
		return tcell.NewEventKey(tcell.KeyCtrlF, 'f', tcell.ModCtrl)
	case tcell.KeyCtrlC:
		ui.out <- "/quit"
		return nil
	default:
		ui.app.SetFocus(ui.userInputPane)
		return event
	}
}

func (ui *UI) redrawConversation() {
	text := make([]string, 0, len(ui.conversation))
	for _, msg := range ui.conversation {
		name, color := "Henk", "green"
		if msg.Type == TypeUser {
			name, color = "You", "blue"
		}
		text = append(text, fmt.Sprintf("[%s]%s[white] %s", color, name, msg.Body))
	}
	ui.conversationPane.SetText(strings.Join(text, "\n")).
		ScrollToEnd()
}
