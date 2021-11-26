package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/net/publicsuffix"
)

const (
	contentTypeJSON = "application/json"

	loginButtonText    = "[ Login ]"
	registerButtonText = "[ Register ]"
	submitButtonText   = "[ Submit ]"
)

var (
	titleStyle   = lipgloss.NewStyle().Bold(true)
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorStyle  = focusedStyle.Copy()
	noStyle      = lipgloss.NewStyle()
	docStyle     = lipgloss.NewStyle().Margin(vertMargin, horizMargin)
)

var (
	httpClient *http.Client
)

type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Text      string    `json:"text"`
	Tags      []string  `json:"tags"`
}

func initialModel(loggedIn bool) tea.Model {
	if !loggedIn {
		return login()
	}
	return home()
}

type mainMenu struct {
	choices    []string
	cursor     int
	windowSize [2]int
}

func home() mainMenu {
	return mainMenu{
		choices: []string{
			"Search log entries",
			"Account Management",
		},
		windowSize: [2]int{0, 0},
	}
}

func (m mainMenu) Init() tea.Cmd {
	return nil
}

func (m mainMenu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter", " ":
			switch m.cursor {
			case 0:
				return search(m.windowSize[0], m.windowSize[1]), nil
			case 1:
				return account(), nil
			}
		}
	case tea.WindowSizeMsg:
		m.windowSize = [2]int{msg.Width, msg.Height}
	}
	return m, nil
}

func (m mainMenu) View() string {
	s := "What do you want to do?\n\n"

	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		s += fmt.Sprintf("%s %s\n", cursor, choice)
	}

	s += "\nPress q to quit.\n"

	return s
}
func main() {
	cookieJar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	httpClient = &http.Client{Jar: cookieJar}
	if err != nil {
		log.Println(err)
		return
	}
	var loggedIn bool
	defer func() {
		r := recover()
		if r != nil {
			log.Println(r)
		}
	}()
	p := tea.NewProgram(initialModel(loggedIn), tea.WithAltScreen())
	if err := p.Start(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
