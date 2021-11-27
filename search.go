package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	vertMargin  = 1
	horizMargin = 2
)

type item struct {
	LogEntry
}

func (i item) Title() string       { return fmt.Sprintf("%s: %s", i.Timestamp.Format(time.RFC3339), i.Text) }
func (i item) Description() string { return strings.Join(i.Tags, ",") }
func (i item) FilterValue() string { return i.Text }

type searchMenu struct {
	showResult bool
	focusIndex int
	inputs     []textinput.Model
	list       list.Model
}

func search(width, height int) searchMenu {
	s := searchMenu{
		inputs: make([]textinput.Model, 3),
		list:   list.NewModel(nil, list.NewDefaultDelegate(), width-(2*horizMargin), height-(2*vertMargin)),
	}

	s.list.Title = "Found Log Entries:"

	var t textinput.Model
	for i := range s.inputs {
		t = textinput.NewModel()
		t.CursorStyle = cursorStyle
		t.CharLimit = 20
		t.SetCursorMode(textinput.CursorStatic)

		switch i {
		case 0:
			t.Placeholder = "Start (2021-06-01T11:22:33Z)"
			t.Focus()
			t.PromptStyle = focusedStyle
			t.TextStyle = focusedStyle
		case 1:
			t.Placeholder = "Start (2021-06-01T11:22:33Z)"
		case 2:
			t.Placeholder = "Tags (comma separated)"
			t.CharLimit = 0
		}

		s.inputs[i] = t
	}
	return s
}

func (s searchMenu) Init() tea.Cmd {
	return nil
}

func (m searchMenu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			if !m.showResult {
				return home(), nil
			}
			m.showResult = false
			return m, nil
		case "tab", "shift+tab", "enter", "up", "down":
			if m.showResult {
				break
			}
			s := msg.String()
			if s == "enter" && m.focusIndex == len(m.inputs) {
				return m, m.getEntries
			}
			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex > len(m.inputs) {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs)
			}

			cmds := make([]tea.Cmd, len(m.inputs))
			for i := 0; i <= len(m.inputs)-1; i++ {
				if i == m.focusIndex {
					cmds[i] = m.inputs[i].Focus()
					m.inputs[i].PromptStyle = focusedStyle
					m.inputs[i].TextStyle = focusedStyle
					continue
				}
				m.inputs[i].Blur()
				m.inputs[i].PromptStyle = noStyle
				m.inputs[i].TextStyle = noStyle
			}
			return m, tea.Batch(cmds...)
		}
	case []LogEntry:
		cmd := m.loadEntries(msg)
		m.list.SetFilteringEnabled(false)
		m.showResult = true
		return m, cmd
	case tea.WindowSizeMsg:
		top, right, bottom, left := docStyle.GetMargin()
		m.list.SetSize(msg.Width-left-right, msg.Height-top-bottom)
	case error:
		log.Println(msg)
	}
	if m.showResult {
		var cmd tea.Cmd

		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}

	return m, m.updateInputs(msg)
}

func (s *searchMenu) updateInputs(msg tea.Msg) tea.Cmd {
	var cmds = make([]tea.Cmd, len(s.inputs))

	for i := range s.inputs {
		s.inputs[i], cmds[i] = s.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (s *searchMenu) getEntries() tea.Msg {
	//TODO: add input validation
	u, err := url.Parse(appConfig.BuildURL("/client/get_entries"))
	if err != nil {
		return err
	}

	start := s.inputs[0].Value()
	end := s.inputs[1].Value()
	tags := s.inputs[2].Value()

	q := u.Query()
	if start != "" {
		q.Add("start", start)
	}
	if end != "" {
		q.Add("end", end)
	}
	if tags != "" {
		q.Add("tags", tags)
	}
	u.RawQuery = q.Encode()

	res, err := httpClient.Get(u.String())
	if err != nil {
		return err
	}
	defer func() {
		res.Body.Close()
	}()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	logEntries := make([]LogEntry, 0)
	err = json.Unmarshal(body, &logEntries)
	if err != nil {
		return err
	}
	return logEntries
}

func (s *searchMenu) loadEntries(entries []LogEntry) tea.Cmd {
	items := make([]list.Item, 0, len(entries))
	for _, entry := range entries {
		items = append(items, item{entry})
	}
	return s.list.SetItems(items)
}

func (s searchMenu) View() string {
	if s.showResult {
		return s.resultView()
	} else {
		return s.menuView()
	}
}

func (s searchMenu) resultView() string {

	return docStyle.Render(s.list.View())
}

func (s searchMenu) menuView() string {
	var b strings.Builder

	for i := range s.inputs {
		b.WriteString(s.inputs[i].View())
		if i < len(s.inputs)-1 {
			b.WriteRune('\n')
		}
	}

	buttonStyle := &blurredStyle
	if s.focusIndex == len(s.inputs) {
		buttonStyle = &focusedStyle
	}
	fmt.Fprintf(&b, "\n\n%s\n\n", buttonStyle.Render(submitButtonText))

	return b.String()
}
