package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type loginPage struct {
	view        loginPageView
	loginForm   loginForm
	confirmForm confirmForm
}

type loginPageView int

const (
	loginPageLogin   loginPageView = iota
	loginPageConfirm loginPageView = iota
)

type loginForm struct {
	focusIndex     loginFormIndex
	urlInput       textinput.Model
	emailInput     textinput.Model
	passwordInput  textinput.Model
	emailErrors    []string
	passwordErrors []string
	Error          error
}

type loginFormIndex int

const (
	loginFormIndexURLInput loginFormIndex = iota
	loginFormIndexEmailInput
	loginFormIndexPasswordInput
	loginFormIndexLoginButton
	loginFormIndexRegisterButton
)

type verificationRequired struct{}

const (
	MessageVerficationRequired = "verification required"
)

var (
	errNoCredentials = errors.New("please enter credentials")
)

func LoginPage() tea.Model {
	return loginPage{
		view:        loginPageLogin,
		loginForm:   LoginForm(),
		confirmForm: ConfirmForm(),
	}
}
func (loginPage) Init() tea.Cmd {
	return nil
}
func (l loginPage) Update(msg tea.Msg) (model tea.Model, cmd tea.Cmd) {
	if l.view == loginPageLogin {
		switch msg.(type) {
		case verificationRequired:
			l.view = loginPageConfirm
			model, cmd = l.confirmForm.Update(msg)
		default:
			model, cmd = l.loginForm.Update(msg)
		}
	} else if l.view == loginPageConfirm {
		model, cmd = l.confirmForm.Update(msg)
	}
	switch model := model.(type) {
	case loginForm:
		l.loginForm = model
	case confirmForm:
		l.confirmForm = model
	default:
		return model, cmd
	}
	return l, cmd
}

func (l loginPage) View() string {
	if l.view == loginPageLogin {
		return l.loginForm.View()
	} else if l.view == loginPageConfirm {
		return l.confirmForm.View()
	}
	panic("login page view out of bounds")
}

func LoginForm() loginForm {
	s := loginForm{}
	var t textinput.Model

	t = textinput.NewModel()
	t.CursorStyle = cursorStyle
	t.Placeholder = "https://mouseion.codemonkeysoftware.net"
	t.Prompt = "URL      > "
	if appConfig.ServerURL != "" {
		t.SetValue(appConfig.ServerURL)
	} else {
		t.Focus()
		t.PromptStyle = focusedStyle
		t.TextStyle = focusedStyle
	}
	s.urlInput = t

	t = textinput.NewModel()
	t.CursorStyle = cursorStyle
	t.Placeholder = "mouseion@example.com"
	t.Prompt = "Email    > "
	if appConfig.ServerURL != "" {
		s.focusIndex = loginFormIndexEmailInput
		t.Focus()
		t.PromptStyle = focusedStyle
		t.TextStyle = focusedStyle
	}
	s.emailInput = t

	t = textinput.NewModel()
	t.CursorStyle = cursorStyle
	t.Placeholder = "Password"
	t.Prompt = "Password > "
	t.EchoMode = textinput.EchoPassword
	t.EchoCharacter = '*'
	s.passwordInput = t

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
				if m.focusIndex == loginFormIndexLoginButton {
					m.emailErrors, m.passwordErrors = nil, nil
					return m, m.login
				}

				if m.focusIndex == loginFormIndexRegisterButton {
					m.emailErrors, m.passwordErrors = nil, nil
					return m, m.register
				}
			}
			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex > loginFormIndexRegisterButton {
				m.focusIndex = loginFormIndexRegisterButton
			} else if m.focusIndex < loginFormIndexURLInput {
				m.focusIndex = loginFormIndexURLInput
			}

			var cmd tea.Cmd
			if m.focusIndex == loginFormIndexURLInput {
				cmd = m.urlInput.Focus()
				m.urlInput.PromptStyle = focusedStyle
				m.urlInput.TextStyle = focusedStyle
			} else {
				m.urlInput.Blur()
				m.urlInput.PromptStyle = noStyle
				m.urlInput.TextStyle = noStyle
			}
			if m.focusIndex == loginFormIndexEmailInput {
				cmd = m.emailInput.Focus()
				m.emailInput.PromptStyle = focusedStyle
				m.emailInput.TextStyle = focusedStyle
			} else {
				m.emailInput.Blur()
				m.emailInput.PromptStyle = noStyle
				m.emailInput.TextStyle = noStyle
			}
			if m.focusIndex == loginFormIndexPasswordInput {
				cmd = m.passwordInput.Focus()
				m.passwordInput.PromptStyle = focusedStyle
				m.passwordInput.TextStyle = focusedStyle
			} else {
				m.passwordInput.Blur()
				m.passwordInput.PromptStyle = noStyle
				m.passwordInput.TextStyle = noStyle
			}
			return m, cmd
		}
	case success:
		return initialModel(true), nil
	case Errors:
		m.emailErrors, m.passwordErrors = msg["email"], msg["password"]
		return m, nil
	case error:
		m.Error = msg
		return m, nil
	case tea.WindowSizeMsg:
		width, height = msg.Width, msg.Height
	}
	cmd := m.updateInputs(msg)
	return m, cmd
}

func (s *loginForm) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, 0, 3)
	var cmd tea.Cmd

	s.emailInput, cmd = s.emailInput.Update(msg)
	cmds = append(cmds, cmd)
	s.urlInput, cmd = s.urlInput.Update(msg)
	cmds = append(cmds, cmd)
	s.passwordInput, cmd = s.passwordInput.Update(msg)
	cmds = append(cmds, cmd)

	return tea.Batch(cmds...)
}

func (s loginForm) View() string {
	var b strings.Builder

	b.WriteString(s.urlInput.View())
	b.WriteRune('\n')

	b.WriteString(s.emailInput.View())
	b.WriteRune('\n')
	if s.emailErrors != nil {
		for _, err := range s.emailErrors {
			fmt.Fprintf(&b, "           %s", errorStyle.Render(err))
			b.WriteRune('\n')
		}
	}

	b.WriteString(s.passwordInput.View())
	b.WriteRune('\n')
	if s.passwordErrors != nil {
		for _, err := range s.passwordErrors {
			fmt.Fprintf(&b, "           %s", errorStyle.Render(err))
			b.WriteRune('\n')
		}
	}
	if s.Error != nil {
		fmt.Fprintf(&b, "\n%s\n", errorStyle.Render(strings.Title(s.Error.Error())))
	}

	buttonStyle := &blurredStyle
	if s.focusIndex == loginFormIndexLoginButton {
		buttonStyle = &focusedStyle
	}
	fmt.Fprintf(&b, "\n%s\n", buttonStyle.Render(loginButtonText))

	buttonStyle = &blurredStyle
	if s.focusIndex == loginFormIndexRegisterButton {
		buttonStyle = &focusedStyle
	}
	fmt.Fprintf(&b, "\n%s\n\n", buttonStyle.Render(registerButtonText))

	return b.String()
}

func (m loginForm) login() tea.Msg {
	appConfig.ServerURL = m.urlInput.Value()
	loginCmd := map[string]string{}

	if m.emailInput.Value() == "" || m.passwordInput.Value() == "" {
		return errNoCredentials
	}

	loginCmd["email"], loginCmd["password"] = m.emailInput.Value(), m.passwordInput.Value()
	b := bytes.NewBuffer(nil)
	encoder := json.NewEncoder(b)
	err := encoder.Encode(loginCmd)
	if err != nil {
		return err
	}

	httpResponse, err := httpClient.Post(appConfig.BuildURL("/auth/login"), contentTypeJSON, b)
	if err != nil {
		return err
	}
	response, err := ParseAPIResponse(httpResponse)
	if err != nil {
		return err
	}

	if response.Message == MessageVerficationRequired || response.Error == MessageVerficationRequired {
		return verificationRequired{}
	}
	if response.Status == StatusSuccess {
		return success{}
	}
	if response.Errors != nil {
		if response.Error != "" {
			response.Errors["error"] = []string{response.Error.Error()}
		}
		return response.Errors
	}
	return response.Error

}

func (m loginForm) register() tea.Msg {
	appConfig.ServerURL = m.urlInput.Value()
	cmd := map[string]string{}

	if m.emailInput.Value() == "" || m.passwordInput.Value() == "" {
		return errNoCredentials
	}

	cmd["email"], cmd["password"], cmd["confirm_password"] = m.emailInput.Value(), m.passwordInput.Value(), m.passwordInput.Value()
	b := bytes.NewBuffer(nil)
	encoder := json.NewEncoder(b)
	err := encoder.Encode(cmd)
	if err != nil {
		return err
	}

	httpResponse, err := httpClient.Post(appConfig.BuildURL("/auth/register"), contentTypeJSON, b)
	if err != nil {
		return err
	}
	response, err := ParseAPIResponse(httpResponse)
	if err != nil {
		return err
	}

	return m.handleResponse(response)
}

func (l loginForm) handleResponse(response *APIResponse) tea.Msg {
	if response.Message == MessageVerficationRequired || response.Error == MessageVerficationRequired {
		return verificationRequired{}
	}
	if response.Status == StatusSuccess {
		return success{}
	}
	if response.Errors != nil {
		if response.Error != "" {
			response.Errors["error"] = []string{response.Error.Error()}
		}
		return response.Errors
	}
	return response.Error
}

type confirmForm struct {
	confirmInput textinput.Model
	Error        error
}

func ConfirmForm() confirmForm {
	form := confirmForm{}
	t := textinput.NewModel()
	t.CursorStyle = cursorStyle
	t.Placeholder = "ABCD123"
	t.Prompt = "Confirm Token > "
	t.Focus()
	t.PromptStyle = focusedStyle
	t.TextStyle = noStyle

	form.confirmInput = t
	return form
}

func (c confirmForm) Init() tea.Cmd {
	return nil
}

func (c confirmForm) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return c, tea.Quit
		case "enter":
			return c, c.confirm
		}
	case success:
		return initialModel(true), nil
	case error:
		c.Error = msg
		return c, nil
	case tea.WindowSizeMsg:
		width, height = msg.Width, msg.Height
	}
	var cmd tea.Cmd
	c.confirmInput, cmd = c.confirmInput.Update(msg)
	return c, cmd
}

func (c confirmForm) View() string {
	var b strings.Builder

	b.WriteString("You must confirm your account before you can continue.\nCheck your email for a confirmation code and enter it below.\n\n")
	b.WriteString(c.confirmInput.View())
	return b.String()
}
func (c confirmForm) confirm() tea.Msg {
	b := &bytes.Buffer{}
	data := map[string]string{"cnf": c.confirmInput.Value()}
	enc := json.NewEncoder(b)
	err := enc.Encode(data)
	if err != nil {
		return fmt.Errorf("encode confirm token: %w", err)
	}
	httpResponse, err := httpClient.Post(appConfig.BuildURL("/auth/confirm"), contentTypeJSON, b)
	if err != nil {
		return err
	}
	response, err := ParseAPIResponse(httpResponse)
	if err != nil {
		return err
	}

	return c.handleResponse(response)
}

func (confirmForm) handleResponse(response *APIResponse) tea.Msg {
	if response.Status == StatusSuccess {
		return success{}
	}
	if response.Errors != nil {
		if response.Error != "" {
			response.Errors["error"] = []string{response.Error.Error()}
		}
		return response.Errors
	}
	return response.Error
}
