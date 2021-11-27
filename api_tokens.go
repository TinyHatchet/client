package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type apiTokenMenu struct {
	cursor  int
	choices []interface{}
}

type deletedID string

const createNewTokenText = "Create New Token"

func APITokenMenu() apiTokenMenu {
	return apiTokenMenu{
		choices: []interface{}{createNewTokenText},
	}
}

func (m apiTokenMenu) Init() tea.Cmd {
	return m.listTokens
}

func (m apiTokenMenu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return account(), nil
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
			if m.cursor == len(m.choices)-1 {
				return m, m.createToken
			}
		case "ctrl+d":
			if m.cursor < len(m.choices)-1 {
				token, ok := m.choices[m.cursor].(apiToken)
				if !ok {
					return m, nil
				}
				return m, m.deleteToken(token)
			}
		}
	case apiToken:
		m.choices[len(m.choices)-1] = msg
		m.choices = append(m.choices, createNewTokenText)
		return m, nil
	case []apiToken:
		if len(msg) == 0 {
			return m, nil
		}
		newChoices := make([]interface{}, 0, len(msg)+1)
		for _, token := range msg {
			newChoices = append(newChoices, token)
		}
		newChoices = append(newChoices, createNewTokenText)
		m.choices = newChoices
		return m, nil
	case deletedID:
		newChoices := make([]interface{}, 0, len(m.choices)-1)
		for _, choice := range m.choices {
			token, ok := choice.(apiToken)
			if ok && token.ID == string(msg) {
				continue
			}
			newChoices = append(newChoices, choice)
		}
		m.choices = newChoices
		return m, nil
	}
	return m, nil
}

func (m apiTokenMenu) View() string {
	b := &strings.Builder{}

	b.WriteString(titleStyle.Render("API Tokens"))
	b.WriteString("\n\n")

	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		switch choice := choice.(type) {
		case apiToken:
			fmt.Fprintf(b, "%s ID: %s\n", cursor, choice.ID)
			if choice.Secret != "" {
				fmt.Fprintf(b, "\tSecret: %s\n", choice.Secret)
			}
		default:
			fmt.Fprintf(b, "%s %s\n", cursor, choice)

		}

	}

	fmt.Fprint(b, "\nPress ctrl+d to delete a token.")
	fmt.Fprint(b, "\nPress q to quit.\n")

	return b.String()
}

type apiToken struct {
	ID     string `json:"id"`
	Secret string `json:"secret"`
}

func (m apiTokenMenu) createToken() tea.Msg {
	resp, err := httpClient.Post(appConfig.BuildURL("/auth/api_token"), contentTypeJSON, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	token := apiToken{}
	err = decoder.Decode(&token)
	if err != nil {
		return err
	}
	return token
}

type listTokensResponse struct {
	Tokens []apiToken `json:"tokens"`
}

func (m apiTokenMenu) listTokens() tea.Msg {
	resp, err := httpClient.Get(appConfig.BuildURL("/auth/api_token"))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	listResponse := listTokensResponse{}
	err = decoder.Decode(&listResponse)
	if err != nil {
		return err
	}
	return listResponse.Tokens
}

func (m apiTokenMenu) deleteToken(token apiToken) tea.Cmd {
	return func() tea.Msg {
		req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf(appConfig.BuildURL("/auth/api_token?id=%s"), token.ID), nil)
		if err != nil {
			return err
		}
		resp, err := httpClient.Do(req)
		if err != nil {
			return err
		}
		if resp.StatusCode != 200 {
			return nil
		}
		return deletedID(token.ID)
	}
}
