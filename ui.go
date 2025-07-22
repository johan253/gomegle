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

// Constants and styles used for splash screen rendering and styling.
const splashMessage = "Welcome to GoMegle"

var (
	gap         = "\n\n" // Space between components
	splashStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
)

// errMsg is used to encapsulate error messages into Bubble Tea Msgs.
type errMsg error

// model defines the state of the Bubble Tea TUI application.
type model struct {
	width           int                // Terminal width
	height          int                // Terminal height
	renderer        *lipgloss.Renderer // Renderer for correct client-side color profile
	splashTimer     timer.Model        // Timer for splash screen
	splashText      string             // Currently displayed splash message
	splashTextIndex int                // Index of next char to append to splashText
	splashSpinner   spinner.Model      // Spinner animation during splash
	viewport        viewport.Model     // Scrollable text window for chat
	messages        []string           // All messages displayed
	textarea        textarea.Model     // Input field for user to type messages
	senderStyle     lipgloss.Style     // Style for user's message prefix
	err             error              // Captured errors
}

// teaHandler wires a Bubble Tea model to a new SSH session.
// This returns the model and Bubble Tea options, such as using the alt screen.
func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	return initialModel(s), []tea.ProgramOption{tea.WithAltScreen()}
}

// initialModel initializes the Bubble Tea model with session-specific settings.
func initialModel(s ssh.Session) model {
	// Setup input box
	ta := textarea.New()
	ta.Placeholder = "Send a message..."
	ta.Focus()
	ta.Prompt = "┃ "
	ta.CharLimit = 280
	ta.SetWidth(30)
	ta.SetHeight(3)
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle() // No line highlighting
	ta.ShowLineNumbers = false
	ta.KeyMap.InsertNewline.SetEnabled(false) // Enter = submit

	// Setup chat display
	vp := viewport.New(30, 5)
	vp.SetContent("Welcome to the chat room!\nType a message and press Enter to send.")

	// Splash screen timer and spinner
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

// Init initializes the Bubble Tea program with starting commands.
func (m model) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, m.splashTimer.Init(), m.splashSpinner.Tick)
}

// Update handles all message types and updates the model accordingly.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		taCmd tea.Cmd
		tiCmd tea.Cmd
		vpCmd tea.Cmd
		ssCmd tea.Cmd
	)

	// Always update textarea and viewport regardless of msg type
	m.textarea, taCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		// Adjust dimensions to fit the terminal
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width
		m.textarea.SetWidth(msg.Width)
		m.viewport.Height = msg.Height - m.textarea.Height() - lipgloss.Height(gap)

		// Rewrap messages when resizing
		if len(m.messages) > 0 {
			m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(strings.Join(m.messages, "\n")))
		}
		m.viewport.GotoBottom()

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			// Exit the program
			fmt.Println(m.textarea.Value())
			return m, tea.Quit
		case "enter":
			// Append message and clear textarea
			m.messages = append(m.messages, m.senderStyle.Render("You: ")+m.textarea.Value())
			m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(strings.Join(m.messages, "\n")))
			m.textarea.Reset()
			m.viewport.GotoBottom()
		}

	case timer.TickMsg:
		// Splash text reveal logic
		if m.splashTextIndex < len(splashMessage) {
			m.splashText += splashMessage[m.splashTextIndex : m.splashTextIndex+1]
			m.splashTextIndex++
		}
		m.splashTimer, tiCmd = m.splashTimer.Update(msg)

	case spinner.TickMsg:
		// Animate splash spinner
		m.splashSpinner, ssCmd = m.splashSpinner.Update(msg)

	case errMsg:
		// Store error
		m.err = msg
		return m, nil
	}

	// Combine all returned commands
	return m, tea.Batch(taCmd, tiCmd, vpCmd, ssCmd)
}

// View renders the entire UI depending on splash state.
func (m model) View() string {
	if !m.splashTimer.Timedout() {
		return splashView(m)
	}
	return fmt.Sprintf("%s%s%s", m.viewport.View(), gap, m.textarea.View())
}

// splashView renders the splash screen animation.
func splashView(m model) string {
	var spinnerText string
	if m.splashTextIndex >= len(splashMessage) {
		spinnerText = m.splashSpinner.View()
	}
	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		fmt.Sprintf("%s %s", m.splashText, splashStyle.Render(spinnerText)),
	)
}
