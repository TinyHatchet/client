package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type success struct{}

type accountMenu struct {
	cursor  int
	choices []string
}

func account() accountMenu {
	return accountMenu{
		choices: []string{"Change Email", "API Tokens"},
	}
}

func (m accountMenu) Init() tea.Cmd {
	return nil
}

func (m accountMenu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return home(), nil
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
				return changeEmail(), nil
			case 1:
				menu := APITokenMenu()
				return menu, menu.Init()
			}
		}
	}
	return m, nil
}

func (m accountMenu) View() string {
	b := &strings.Builder{}

	b.WriteString(titleStyle.Render("Account Management"))
	b.WriteString("\n\n")

	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		fmt.Fprintf(b, "%s %s\n", cursor, choice)
	}

	fmt.Fprint(b, "\nPress q to quit.\n")

	return b.String()
}

type titledInput struct {
	Title string
	textinput.Model
}

type changeEmailForm struct {
	focusIndex int
	inputs     []titledInput
}

func changeEmail() changeEmailForm {
	s := changeEmailForm{inputs: make([]titledInput, 1)}

	for i := range s.inputs {
		t := titledInput{}
		t.Model = textinput.NewModel()
		t.CursorStyle = cursorStyle
		t.CharLimit = 0
		t.SetCursorMode(textinput.CursorStatic)

		switch i {
		case 0:
			t.Title = "New Email"
			t.Placeholder = "mouseion@example.com"
			t.Focus()
			t.PromptStyle = focusedStyle
			t.TextStyle = focusedStyle
		}
		s.inputs[i] = t
	}

	return s
}

func (form changeEmailForm) Init() tea.Cmd {
	return nil
}

func (form changeEmailForm) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return form, tea.Quit
		case "esc":
			return account(), nil
		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()
			if s == "enter" && form.focusIndex == len(form.inputs) {
				return form, form.updateEmail
			}
			if s == "up" || s == "shift+tab" {
				form.focusIndex--
			} else {
				form.focusIndex++
			}

			if form.focusIndex > len(form.inputs) {
				form.focusIndex = 0
			} else if form.focusIndex < 0 {
				form.focusIndex = len(form.inputs)
			}

			cmds := make([]tea.Cmd, len(form.inputs))
			for i := 0; i <= len(form.inputs)-1; i++ {
				if i == form.focusIndex {
					cmds[i] = form.inputs[i].Focus()
					form.inputs[i].PromptStyle = focusedStyle
					form.inputs[i].TextStyle = focusedStyle
					continue
				}
				form.inputs[i].Blur()
				form.inputs[i].PromptStyle = noStyle
				form.inputs[i].TextStyle = noStyle
			}
			return form, tea.Batch(cmds...)
		}
	case success:
		return account(), nil
	}
	return form, form.updateInputs(msg)
}

func (form *changeEmailForm) updateInputs(msg tea.Msg) tea.Cmd {
	var cmds = make([]tea.Cmd, len(form.inputs))

	for i := range form.inputs {

		form.inputs[i].Model, cmds[i] = form.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}
func (form changeEmailForm) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Account Management"))
	b.WriteString("\n\n")
	for i := range form.inputs {
		b.WriteString(form.inputs[i].Title)
		b.WriteRune(' ')
		b.WriteString(form.inputs[i].View())
		if i < len(form.inputs)-1 {
			b.WriteRune('\n')
		}
	}

	buttonStyle := &blurredStyle
	if form.focusIndex == len(form.inputs) {
		buttonStyle = &focusedStyle
	}
	fmt.Fprintf(&b, "\n\n%s\n\n", buttonStyle.Render(submitButtonText))

	return b.String()
}

type changeEmailCommand struct {
	Email string `json:"email"`
}

func (form changeEmailForm) updateEmail() tea.Msg {
	cmd := changeEmailCommand{
		Email: form.inputs[0].Value(),
	}
	buf := bytes.NewBuffer(nil)
	encoder := json.NewEncoder(buf)
	err := encoder.Encode(cmd)
	if err != nil {
		return err
	}
	resp, err := httpClient.Post(appConfig.BuildURL("/account/change_email"), contentTypeJSON, buf)
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusOK {
		return success{}
	}
	//TODO: Handle Errors
	return nil
}
