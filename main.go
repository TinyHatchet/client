package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/net/publicsuffix"

	"gopkg.in/yaml.v2"
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
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF44475A"))
	cursorStyle  = focusedStyle.Copy()
	noStyle      = lipgloss.NewStyle()
	docStyle     = lipgloss.NewStyle().Margin(vertMargin, horizMargin)
)

var (
	httpClient *http.Client
	width      int
	height     int
)

type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Text      string    `json:"text"`
	Tags      []string  `json:"tags"`
}

func initialModel(loggedIn bool) tea.Model {
	if !loggedIn {
		return LoginPage()
	}
	return home()
}

type mainMenu struct {
	choices []string
	cursor  int
}

func home() mainMenu {
	return mainMenu{
		choices: []string{
			"Search log entries",
			"Account Management",
		},
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
				return search(), nil
			case 1:
				return account(), nil
			}
		}
	case tea.WindowSizeMsg:
		width, height = msg.Width, msg.Height
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

var appConfig Config

type Config struct {
	ServerURL    string
	EmailAddress string
	DebugPath    string
}

func (c *Config) LoadFromFile(path string) error {
	body, err := ioutil.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read config file: %w", err)
	}
	err = yaml.Unmarshal(body, c)
	if err != nil {
		return fmt.Errorf("unmarshal config: %w", err)
	}
	c.ServerURL = "https://tinyhatchet.com"
	return nil
}

func (c Config) WriteOut(path string) error {
	body, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	err = ioutil.WriteFile(path, body, 0600)
	if err != nil {
		return fmt.Errorf("write config file %w", err)
	}
	return nil
}

func (c Config) BuildURL(path string) string {
	return fmt.Sprintf("%s%s", c.ServerURL, path)
}

func main() {
	homedir, _ := os.UserHomeDir()
	defaultConfig := homedir + string(os.PathSeparator) + ".tinyhatchet.config"
	var configPath string
	flag.StringVar(&configPath, "config", defaultConfig, "")
	flag.Parse()

	err := appConfig.LoadFromFile(configPath)
	if err != nil {
		log.Fatal(err)
	}

	defer appConfig.WriteOut(configPath)

	cookieJar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	httpClient = &http.Client{Jar: cookieJar}
	if err != nil {
		log.Fatal(err)
	}
	var loggedIn bool
	defer func() {
		r := recover()
		if r != nil {
			log.Println(r)
		}
	}()

	if appConfig.DebugPath != "" {
		_, _ = tea.LogToFile("debug.log", "")
	}

	p := tea.NewProgram(initialModel(loggedIn), tea.WithAltScreen())
	if err := p.Start(); err != nil {
		log.Fatalf("Alas, there's been an error: %v", err)
	}
}
