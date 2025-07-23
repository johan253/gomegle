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
	user            *User              // User channels for sending/receiving
	matched         bool               // Whether user has been matched
	promptRequeue   bool               // Whether to prompt user to requeue after getting unmatched
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
	vp.SetContent("Welcome to GoMegle!\nLooking for someone to chat with...")

	// Splash screen timer and spinner
	timer := timer.NewWithInterval(2*time.Second, 30*time.Millisecond)
	r := bubbletea.MakeRenderer(s)

	ss := spinner.New()
	ss.Spinner = spinner.Points

	// Create user with channels and add to matchmaker
	user := &User{
		receive: make(chan ChatMsg, 100), // Buffered to prevent blocking
		send:    nil,                     // Will be set when matched
	}

	// Add user to matchmaker queue
	globalMatchmaker.Enqueue(user)

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
		user:            user,
		matched:         false,
		promptRequeue:   false, // Initially not prompting for requeue
	}
}

// Init initializes the Bubble Tea program with starting commands.
func (m model) Init() tea.Cmd {
	return tea.Batch(
		textarea.Blink,
		m.splashTimer.Init(),
		m.splashSpinner.Tick,
		m.user.ListenForMessages(), // Start listening for messages
	)
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
		case "ctrl+c":
			// If the user is in the queue, remove them and close thier channel
			if !m.matched {
				globalMatchmaker.Dequeue(m.user)
			} else {
				// If matched, send leave message
				leaveMsg := ChatMsg{
					Type:    ChatMsgTypeLeave,
					Content: "Stranger has left the chat.",
				}
				m.user.send <- leaveMsg
			}
			close(m.user.receive)
			// Exit the program
			return m, tea.Quit
		case "enter":
			// Send message if matched and textarea has content
			if m.matched && m.user.send != nil && m.textarea.Value() != "" {
				chatMsg := ChatMsg{
					Type:    ChatMsgTypeMessage,
					Content: m.textarea.Value(),
				}

				// Send message to other user (non-blocking)
				select {
				case m.user.send <- chatMsg:
					// Message sent successfully, add to our view
					m.messages = append(m.messages, m.senderStyle.Render("You: ")+m.textarea.Value())
				default:
					// Channel is full or closed, show error
					m.messages = append(m.messages, "Error: Could not send message")

				}
			} else if m.promptRequeue && m.textarea.Value() == "r" {
				// If prompted to requeue, re-add user to matchmaker
				globalMatchmaker.Enqueue(m.user)
				m.promptRequeue = false
				m.messages = append(m.messages, "Requeued! Waiting for a new match...")

			}
			m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(strings.Join(m.messages, "\n")))
			m.viewport.GotoBottom()
			m.textarea.Reset()
		}

	case chatMsgReceived:
		// Handle received messages from other users
		switch msg.Type {
		case ChatMsgTypeJoin:
			m.matched = true
			m.promptRequeue = false
			m.messages = append(m.messages, "✅ "+msg.Content)
		case ChatMsgTypeMessage:
			m.messages = append(m.messages, "Stranger: "+msg.Content)
		case ChatMsgTypeLeave:
			m.matched = false
			m.promptRequeue = true
			m.messages = append(m.messages, "❌ "+msg.Content)
			m.messages = append(m.messages, "Send 'r' to requeue or press 'ctrl+c' to exit.")
		case ChatMsgTypeError:
			m.messages = append(m.messages, "🚨 "+msg.Content)
		}

		// Update viewport and scroll to bottom
		m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(strings.Join(m.messages, "\n")))
		m.viewport.GotoBottom()

		// Continue listening for more messages
		return m, m.user.ListenForMessages()

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

	// Show connection status in the input placeholder
	if m.matched {
		m.textarea.Placeholder = "Type your message..."
	} else {
		m.textarea.Placeholder = "Waiting for match..."
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
