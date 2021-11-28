package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type loginForm struct {
	focusIndex int
	inputs     []titledInput
}

func login() loginForm {
	s := loginForm{
		inputs: make([]titledInput, 3),
	}

	for i := range s.inputs {
		var t titledInput
		t.Model = textinput.NewModel()
		t.CursorStyle = cursorStyle
		t.SetCursorMode(textinput.CursorStatic)

		switch i {
		case 0:
			t.Title = "Server URL"
			t.Placeholder = "https://mouseion.codemonkeysoftware.net"
			if appConfig.ServerURL != "" {
				t.SetValue(appConfig.ServerURL)
			} else {
				t.Focus()
				t.PromptStyle = focusedStyle
				t.TextStyle = focusedStyle
			}
		case 1:
			t.Title = "Email"
			t.Placeholder = "mouseion@example.com"
			if appConfig.ServerURL != "" {
				s.focusIndex = 1
				t.Focus()
				t.PromptStyle = focusedStyle
				t.TextStyle = focusedStyle
			}
		case 2:
			t.Title = "Password"
			t.Placeholder = "Password"
			t.EchoMode = textinput.EchoPassword
			t.EchoCharacter = '*'
		}

		s.inputs[i] = t
	}
	return s
}

func (s loginForm) Init() tea.Cmd {
	return nil
}

func (m loginForm) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()
			if s == "enter" {
				if m.focusIndex == len(m.inputs) {
					return m, m.login
				}

				if m.focusIndex == len(m.inputs)+1 {
					return m, m.register
				}
			}
			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex > len(m.inputs)+1 {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs) + 1
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
	case success:
		return initialModel(true), nil

	case tea.WindowSizeMsg:
		width, height = msg.Width, msg.Height
	}

	return m, m.updateInputs(msg)
}

func (s *loginForm) updateInputs(msg tea.Msg) tea.Cmd {
	var cmds = make([]tea.Cmd, len(s.inputs))

	for i := range s.inputs {
		s.inputs[i].Model, cmds[i] = s.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (s loginForm) View() string {
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
	fmt.Fprintf(&b, "\n\n%s\n", buttonStyle.Render(loginButtonText))

	buttonStyle = &blurredStyle
	if s.focusIndex == len(s.inputs)+1 {
		buttonStyle = &focusedStyle
	}
	fmt.Fprintf(&b, "\n%s\n\n", buttonStyle.Render(registerButtonText))

	return b.String()
}

func (m loginForm) login() tea.Msg {
	appConfig.ServerURL = m.inputs[0].Value()
	loginCmd := map[string]string{}
	for _, v := range m.inputs {
		if v.Value() == "" {
			return nil
		}
	}
	loginCmd["email"], loginCmd["password"] = m.inputs[1].Value(), m.inputs[2].Value()
	b := bytes.NewBuffer(nil)
	encoder := json.NewEncoder(b)
	err := encoder.Encode(loginCmd)
	if err != nil {
		return err
	}

	response, err := httpClient.Post(appConfig.BuildURL("/auth/login"), contentTypeJSON, b)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != 307 {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return err
		}
		return body
	}

	return success{}
}

func (m loginForm) register() tea.Msg {
	return success{}
}
