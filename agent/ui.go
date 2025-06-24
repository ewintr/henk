package agent

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
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
}

func NewUI(cancel context.CancelFunc) *UI {
	ui := &UI{
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
	go ui.processInput()
	ui.out <- "ui ready"

	ui.captureInput()
}

func (ui *UI) processInput() {
	for msg := range ui.in {
		switch msg.Type {
		case TypeGeneral:
			fmt.Printf("\033[37mStatus: %s\033[0m\n", msg.Body)
		case TypeHenk:
			ui.conversation = append(ui.conversation, msg)
			fmt.Printf("\033[32mHenk: %s\033[0m\n", msg.Body)
		case TypePrompt:
			fmt.Printf("> ")
		case TypeUser:
			fmt.Printf("\033[34mYou: %s\033[0m\n", msg.Body)
		case TypeTool:
			fmt.Printf("\033[33mTool: %s\033[0m\n", msg.Body)
		case TypeError:
			fmt.Printf("\033[31mError: %s\033[0m\n", msg.Body)
		case TypeDebug:
			fmt.Printf("\033[90mDebug: %s\033[0m\n", msg.Body)
		case TypeExit:
			ui.Close()
		}
	}
}

func (ui *UI) Close() {
	ui.cancel()
	close(ui.in)
	close(ui.out)
}

func (ui *UI) captureInput() {
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		// Add user message to conversation if not a command
		if !strings.HasPrefix(input, "/") {
			msg := Message{
				Type: TypeUser,
				Body: input,
			}
			ui.conversation = append(ui.conversation, msg)
		}

		// Send input to agent
		ui.out <- input

	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading input: %v\n", err)
	}
}
