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
	gossh "golang.org/x/crypto/ssh"
)

// Constants and styles used for splash screen rendering and styling.
const (
	splashMessage  = "Welcome to GoMegle"
	welcomeMessage = "Welcome to GoMegle!\nSend '\\h' at any time to open help menu.\nLooking for someone to chat with..."
)

// TODO: Implement auto-requeue
const helpText = `
\h     - Show this help menu
\q     - Disconnect from current chat
\r     - Requeue for a new chat
\a 	   - Toggle auto-requeue [COMING SOON]

q      - Exit this help menu
ctrl+c - Exit the app at any time
`

var (
	gap         = "\n\n" // Space between components
	splashStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
)

// errMsg is used to encapsulate error messages into Bubble Tea Msgs.
type errMsg error

type UIState int

const (
	StateUIMenu UIState = iota
	StateUIChat
	StateUIHelp
)

type ChatState int

const (
	StateChatMatched ChatState = iota
	StateChatQueued
	StateChatDisconnected
)

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
	helpMenu        bool               // Whether to show help menu
	uiState         UIState            // Current state of the UI
	chatState       ChatState          // Current state of the chat
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
	// vp.SetContent("Welcome to GoMegle!\nLooking for someone to chat with...")

	// Splash screen timer and spinner
	timer := timer.NewWithInterval(2*time.Second, 30*time.Millisecond)
	r := bubbletea.MakeRenderer(s)

	ss := spinner.New()
	ss.Spinner = spinner.Points

	pk := string(gossh.MarshalAuthorizedKey(s.PublicKey()))

	// Create user with channels and add to matchmaker
	user := &User{
		pubKey:  pk,
		receive: make(chan ChatMsg, 100), // Buffered to prevent blocking
		send:    nil,                     // Will be set when matched
	}

	// Add user to matchmaker queue
	// globalMatchmaker.Enqueue(user)

	return model{
		width:           30,
		height:          10,
		renderer:        r,
		splashTimer:     timer,
		splashText:      "",
		splashTextIndex: 0,
		splashSpinner:   ss,
		textarea:        ta,
		messages:        []string{welcomeMessage},
		viewport:        vp,
		senderStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		user:            user,
		matched:         false,
		promptRequeue:   false, // Initially not prompting for requeue
		helpMenu:        false, // Initially not showing help menu
		uiState:         StateUIMenu,
		chatState:       StateChatDisconnected,
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
	case timer.TimeoutMsg:
		m.uiState = StateUIChat
		// Enqueue the user after the splash screen times out
		if m.chatState != StateChatMatched {
			globalMatchmaker.Enqueue(m.user)
			m.chatState = StateChatQueued
		}
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
		return m.handleKeyMsg(msg)
	case chatMsgReceived:
		// Handle received messages from other users
		switch msg.Type {
		case ChatMsgTypeJoin:
			m.chatState = StateChatMatched
			m.matched = true
			m.promptRequeue = false
			m.messages = append(m.messages, "✅ "+msg.Content)
		case ChatMsgTypeMessage:
			m.messages = append(m.messages, "Stranger: "+msg.Content)
		case ChatMsgTypeLeave:
			m.chatState = StateChatDisconnected
			m.matched = false
			m.promptRequeue = true
			m.messages = append(m.messages, "❌ "+msg.Content)
			m.messages = append(m.messages, "Send '\\r' to requeue or press 'ctrl+c' to exit.")
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

func (m model) handleKeyMsg(msg tea.KeyMsg) (model, tea.Cmd) {
	key := msg.String()
	// global keybind ctrl+c to exit
	if key == "ctrl+c" {
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
	}
	// Handle key messages based on current state
	switch m.uiState {
	// In the chat UI
	case StateUIChat:
		switch m.chatState {
		// Currently matched with another user
		case StateChatMatched:
			switch key {
			case "enter":
				switch strings.TrimSpace(m.textarea.Value()) {
				case "":
				case "\\h":
					m.uiState = StateUIHelp
				case "\\q":
					leaveMsg := ChatMsg{
						Type:    ChatMsgTypeLeave,
						Content: "Stranger has left the chat.",
					}
					m.user.send <- leaveMsg
					m.chatState = StateChatDisconnected
					m.messages = append(m.messages, "You have left the chat. Send '\\r' to requeue or press 'ctrl+c' to exit.")
				default:
					chatMsg := ChatMsg{
						Type:    ChatMsgTypeMessage,
						Content: strings.TrimSpace(m.textarea.Value()),
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
				}
			}
		// Not matched, and not in queue
		case StateChatDisconnected:
			switch key {
			case "enter":
				switch m.textarea.Value() {
				case "\\h":
					m.uiState = StateUIHelp
				case "\\r":
					globalMatchmaker.Enqueue(m.user)
					m.chatState = StateChatQueued
					m.messages = append(m.messages, "Requeued! Send '\\q' to exit queue. Waiting for a new match...")
				}
			}
		// Not matched, but in queue
		case StateChatQueued:
			switch key {
			case "enter":
				switch m.textarea.Value() {
				case "\\h":
					m.uiState = StateUIHelp
				case "\\q":
					globalMatchmaker.Dequeue(m.user)
					m.chatState = StateChatDisconnected
					m.messages = append(m.messages, "You have left the queue. Send '\\r' to requeue or press 'ctrl+c' to exit.")
				}
			}
		}
	// In the menu UI
	case StateUIMenu:
	// In the help UI
	case StateUIHelp:
		if key == "q" {
			m.uiState = StateUIChat
			m.textarea.Reset() // Reset textarea when exiting help
		}
	}
	// global keybinds below, after handling state-specific keybinds
	if key == "enter" {
		m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(strings.Join(m.messages, "\n")))
		m.viewport.GotoBottom()
		m.textarea.Reset()
	}
	return m, nil
}

// View renders the entire UI depending on model state.
func (m model) View() string {
	if !m.splashTimer.Timedout() {
		return splashView(m)
	}
	if m.uiState == StateUIHelp {
		return helpView(m)
	}

	// Show connection status in the input placeholder
	if m.chatState == StateChatMatched {
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

func helpView(m model) string {
	return m.renderer.NewStyle().
		Padding(1, 1).
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Left, lipgloss.Center).
		Render(helpText)
}
