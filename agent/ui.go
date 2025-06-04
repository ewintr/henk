package agent

import (
	"bufio"
	"fmt"
	"os"
)

type MessageType string

const (
	TypeGeneral MessageType = "general"
	TypeHenk    MessageType = "henk"
	TypePrompt  MessageType = "prompt"
	TypeTool    MessageType = "tool"
	TypeError   MessageType = "error"
	TypeExit    MessageType = "exit"
)

type Message struct {
	Type MessageType
	Body string
}

type UI struct {
	in  chan Message
	out chan string
}

func NewUI() *UI {
	ui := &UI{
		in:  make(chan Message),
		out: make(chan string),
	}
	go ui.Run()

	return ui
}

func (ui *UI) In() chan Message { return ui.in }
func (ui *UI) Out() chan string { return ui.out }

func (ui *UI) Run() {
	scanner := bufio.NewScanner(os.Stdin)
	for msg := range ui.in {
		switch msg.Type {
		case TypeGeneral:
			fmt.Println(msg.Body)
		case TypeHenk:
			fmt.Printf("\u001b[93mHenk\u001b[0m: %s\n", msg.Body)
		case TypeTool:
			fmt.Printf("\u001b[92mtool\u001b[0m: %s\n", msg.Body)
		case TypePrompt:
			fmt.Print("\u001b[94mYou\u001b[0m: ")
			if !scanner.Scan() {
				ui.in <- Message{
					Type: TypeError,
					Body: "Could not read user input.",
				}
			}
			userInput := scanner.Text()
			ui.out <- userInput
		case TypeError:
			fmt.Printf("\u001b[91mError\u001b[0m: %s\n", msg.Body)
		case TypeExit:
			fmt.Println("Bye!")
			ui.Close()
		}
	}
}

func (ui *UI) Close() {
	close(ui.in)
	close(ui.out)
}
