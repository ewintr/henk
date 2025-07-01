package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/briandowns/spinner"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/huh"
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
	conversation []Message
	in           chan Message
	out          chan string
	cancel       context.CancelFunc
	spinner      *spinner.Spinner
}

func NewUI(cancel context.CancelFunc) *UI {
	sp := spinner.New([]string{"  .  ", "  .. ", "  ..."}, 500*time.Millisecond)
	sp.FinalMSG = ""

	ui := &UI{
		in:      make(chan Message),
		out:     make(chan string),
		cancel:  cancel,
		spinner: sp,
	}
	go ui.run()

	return ui
}

func (ui *UI) In() chan Message { return ui.in }
func (ui *UI) Out() chan string { return ui.out }

func (ui *UI) run() {
	ui.out <- "ui ready"
	for msg := range ui.in {
		ui.spinner.Stop()

		if msg.Type == TypePrompt {
			var result string
			huh.NewText().
				CharLimit(400).
				Value(&result).
				Run()
			ui.out <- result
			msg = Message{Type: TypeUser, Body: result}
		}

		var who string
		switch msg.Type {
		case TypeGeneral:
			who = "Agent"
		case TypeHenk:
			ui.conversation = append(ui.conversation, msg)
			who = "Henk"
		case TypeUser:
			who = "You"
		case TypeTool:
			who = "Tool"
		case TypeError:
			who = "Error"
		case TypeDebug:
			who = "Debug"
		case TypeExit:
			ui.Close()
			return
		}

		in := fmt.Sprintf("**%s**: %s", who, msg.Body)
		out, err := glamour.Render(in, "dark")
		if err != nil {
			fmt.Println(err)
		}
		fmt.Print(out)

		ui.spinner.Start()
	}
}

func (ui *UI) Close() {
	if ui.spinner.Active() {
		ui.spinner.Stop()
	}
	ui.cancel()
	close(ui.in)
	close(ui.out)
}
