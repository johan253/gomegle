package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/timer"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish/bubbletea"
)

const splashMessage = "Welcome to GoMegle"

var (
	gap         = "\n\n"
	splashStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
)

type errMsg error

type model struct {
	width           int
	height          int
	renderer        *lipgloss.Renderer
	splashTimer     timer.Model
	splashText      string
	splashTextIndex int
	splashSpinner   spinner.Model
	viewport        viewport.Model
	messages        []string
	textarea        textarea.Model
	senderStyle     lipgloss.Style
	err             error
}

func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	return initialModel(s), []tea.ProgramOption{tea.WithAltScreen()}
}

func initialModel(s ssh.Session) model {
	ta := textarea.New()
	ta.Placeholder = "Send a message..."
	ta.Focus()
	ta.Prompt = "┃ "
	ta.CharLimit = 280
	ta.SetWidth(30)
	ta.SetHeight(3)
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.ShowLineNumbers = false
	ta.KeyMap.InsertNewline.SetEnabled(false)

	vp := viewport.New(30, 5)
	vp.SetContent("Welcome to the chat room!\nType a message and press Enter to send.")

	timer := timer.NewWithInterval(2*time.Second, 30*time.Millisecond)
	r := bubbletea.MakeRenderer(s)

	ss := spinner.New()
	ss.Spinner = spinner.Points

	return model{
		width:           30,
		height:          10,
		renderer:        r,
		splashTimer:     timer,
		splashText:      "",
		splashTextIndex: 0,
		splashSpinner:   ss,
		textarea:        ta,
		messages:        []string{},
		viewport:        vp,
		senderStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, m.splashTimer.Init(), m.splashSpinner.Tick)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		taCmd tea.Cmd
		tiCmd tea.Cmd
		vpCmd tea.Cmd
		ssCmd tea.Cmd
	)

	m.textarea, taCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width
		m.textarea.SetWidth(msg.Width)
		m.viewport.Height = msg.Height - m.textarea.Height() - lipgloss.Height(gap)

		if len(m.messages) > 0 {
			m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(strings.Join(m.messages, "\n")))
		}
		m.viewport.GotoBottom()

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			fmt.Println(m.textarea.Value())
			return m, tea.Quit
		case "enter":
			m.messages = append(m.messages, m.senderStyle.Render("You: ")+m.textarea.Value())
			m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(strings.Join(m.messages, "\n")))
			m.textarea.Reset()
			m.viewport.GotoBottom()
		}

	case timer.TickMsg:
		if m.splashTextIndex < len(splashMessage) {
			m.splashText += splashMessage[m.splashTextIndex : m.splashTextIndex+1]
			m.splashTextIndex++
		}
		m.splashTimer, tiCmd = m.splashTimer.Update(msg)

	case spinner.TickMsg:
		m.splashSpinner, ssCmd = m.splashSpinner.Update(msg)

	case errMsg:
		m.err = msg
		return m, nil
	}

	return m, tea.Batch(taCmd, tiCmd, vpCmd, ssCmd)
}

func (m model) View() string {
	if !m.splashTimer.Timedout() {
		return splashView(m)
	}
	return fmt.Sprintf("%s%s%s", m.viewport.View(), gap, m.textarea.View())
}

func splashView(m model) string {
	var spinnerText string
	if m.splashTextIndex >= len(splashMessage) {
		spinnerText = m.splashSpinner.View()
	}
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
		fmt.Sprintf("%s %s", m.splashText, splashStyle.Render(spinnerText)))
}
