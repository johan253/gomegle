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

const helpText = `
\h     - Show this help menu
\q     - Disconnect from current chat, or queue
\r     - Requeue for a new chat
\a 	   - Toggle auto-requeue
\c     - Clear chat window

q      - Exit this help menu
ctrl+c - Exit the app at any time
`

var gap = "\n\n" // Space between components

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
	splashStyle     lipgloss.Style     // Style for splash text
	viewport        viewport.Model     // Scrollable text window for chat
	messages        []string           // All messages displayed
	textarea        textarea.Model     // Input field for user to type messages
	senderStyle     lipgloss.Style     // Style for user's message prefix
	receiverStyle   lipgloss.Style     // Style for stranger's message prefix
	err             error              // Captured errors
	user            *User              // User channels for sending/receiving
	uiState         UIState            // Current state of the UI
	chatState       ChatState          // Current state of the chat
	autoRequeue     bool               // Whether to auto-requeue after disconnect
	incrFailed      bool               // Whether incrementing active users failed
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
	switchTextAreaStyle(&ta, r)               // Apply renderer styles to textarea
	vp.Style = r.NewStyle().Inherit(vp.Style) // Apply renderer styles to viewport

	ss := spinner.New()
	ss.Spinner = spinner.Dot

	pk := string(gossh.MarshalAuthorizedKey(s.PublicKey()))

	// Create user with channels and add to matchmaker
	pubsub := rdb.Subscribe(ctx, "user:"+pk)
	ch := pubsub.Channel()
	user := &User{
		pubKey:  pk,
		pubsub:  pubsub,
		receive: ch, // Buffered to prevent blocking
	}
	// Add user to matchmaker queue
	// globalMatchmaker.Enqueue(user)
	// Increment player count
	_, err := rdb.Incr(ctx, "active").Result()
	incrFailed := err != nil

	return model{
		width:           30,
		height:          10,
		renderer:        r,
		splashTimer:     timer,
		splashText:      "",
		splashTextIndex: 0,
		splashSpinner:   ss,
		splashStyle:     r.NewStyle().Foreground(lipgloss.Color("3")),
		textarea:        ta,
		messages:        []string{welcomeMessage},
		viewport:        vp,
		senderStyle:     r.NewStyle().Foreground(lipgloss.Color("5")),
		receiverStyle:   r.NewStyle().Foreground(lipgloss.Color("3")),
		user:            user,
		uiState:         StateUIMenu,
		chatState:       StateChatDisconnected,
		autoRequeue:     false, // Auto-requeue disabled by default
		incrFailed:      incrFailed,
	}
}

// switchTextAreaStyle switches the textarea styles to use the renderer's styles.
func switchTextAreaStyle(ta *textarea.Model, r *lipgloss.Renderer) {
	// Apply renderer to cursor
	ta.Cursor.Style = r.NewStyle().Inherit(ta.Cursor.Style)
	ta.Cursor.TextStyle = r.NewStyle().Inherit(ta.Cursor.TextStyle)
	// Apply renderer to focused styles
	ta.FocusedStyle.Base = r.NewStyle().Inherit(ta.FocusedStyle.Base)
	ta.FocusedStyle.CursorLine = r.NewStyle().Inherit(ta.FocusedStyle.CursorLine)
	ta.FocusedStyle.Placeholder = r.NewStyle().Inherit(ta.FocusedStyle.Placeholder)
	ta.FocusedStyle.CursorLineNumber = r.NewStyle().Inherit(ta.FocusedStyle.CursorLineNumber)
	ta.FocusedStyle.EndOfBuffer = r.NewStyle().Inherit(ta.FocusedStyle.EndOfBuffer)
	ta.FocusedStyle.Placeholder = r.NewStyle().Inherit(ta.FocusedStyle.Placeholder)
	ta.FocusedStyle.Prompt = r.NewStyle().Inherit(ta.FocusedStyle.Prompt)
	ta.FocusedStyle.Text = r.NewStyle().Inherit(ta.FocusedStyle.Text)
	// Apply renderer to blurred styles
	ta.BlurredStyle.Base = r.NewStyle().Inherit(ta.BlurredStyle.Base)
	ta.BlurredStyle.CursorLine = r.NewStyle().Inherit(ta.BlurredStyle.CursorLine)
	ta.BlurredStyle.Placeholder = r.NewStyle().Inherit(ta.BlurredStyle.Placeholder)
	ta.BlurredStyle.CursorLineNumber = r.NewStyle().Inherit(ta.BlurredStyle.CursorLineNumber)
	ta.BlurredStyle.EndOfBuffer = r.NewStyle().Inherit(ta.BlurredStyle.EndOfBuffer)
	ta.BlurredStyle.Placeholder = r.NewStyle().Inherit(ta.BlurredStyle.Placeholder)
	ta.BlurredStyle.Prompt = r.NewStyle().Inherit(ta.BlurredStyle.Prompt)
	ta.BlurredStyle.Text = r.NewStyle().Inherit(ta.BlurredStyle.Text)
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
			if err := globalMatchmaker.Enqueue(m.user); err == nil {
				m.chatState = StateChatQueued
			} else {
				m.chatState = StateChatDisconnected
				m.messages = append(m.messages, "Error: Could not enqueue. Try again later.")
			}
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
			m.user.send = msg.Content // Set the other user's public key
			m.messages = append(m.messages, "✅ You matched with a stranger, say hello!")
		case ChatMsgTypeMessage:
			m.messages = append(m.messages, m.receiverStyle.Render("Stranger: ")+msg.Content)
		case ChatMsgTypeLeave:
			m.chatState = StateChatDisconnected
			m.user.send = "" // Clear send channel
			m.messages = append(m.messages, "❌ "+msg.Content)
			if m.autoRequeue {
				if err := globalMatchmaker.Enqueue(m.user); err == nil {
					m.chatState = StateChatQueued
					m.messages = append(m.messages, "Auto-requeue enabled! Waiting for a new match...")
				} else {
					m.messages = append(m.messages, "Error: Could not auto-requeue. Try again later.")
				}
			} else {
				m.messages = append(m.messages, "Send '\\r' to requeue or press 'ctrl+c' to exit.")
			}
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
		if m.chatState == StateChatQueued {
			if err := globalMatchmaker.Dequeue(m.user); err != nil {
				fmt.Printf("Error dequeuing user: %v\n", err)
			}
		} else {
			// If matched, send leave message
			if err := m.user.LeaveChat(); err != nil {
				fmt.Printf("Error leaving chat: %v\n", err)
			}
		}
		if err := m.user.pubsub.Close(); err != nil { // Close the user's pubsub channel
			fmt.Printf("Error closing pubsub: %v\n", err)
		}
		if !m.incrFailed {
			rdb.Decr(ctx, "active") // Decrement active users
		}
		// Exit the program
		return m, tea.Quit
	}
	// Handle key messages based on current state
	switch m.uiState {
	// In the chat UI
	case StateUIChat:
		// handle global keybinds when in chat UI, regardless of chat state
		switch key {
		case "enter":
			switch strings.TrimSpace(m.textarea.Value()) {
			case "":
			case "\\h":
				m.uiState = StateUIHelp
			case "\\c":
				var status string
				switch m.chatState {
				case StateChatMatched:
					status = "in a chat!"
				case StateChatQueued:
					status = "queued!"
				case StateChatDisconnected:
					status = "disconnected!"
				}
				m.messages = []string{"Chat Cleared. Currently " + status}
			case "\\r":
				switch m.chatState {
				case StateChatDisconnected:
					if err := globalMatchmaker.Enqueue(m.user); err == nil {
						m.chatState = StateChatQueued
						m.messages = append(m.messages, "Re-queued! send '\\q' to exit queue or 'ctrl+c' to quit.")
					} else {
						m.messages = append(m.messages, "Error: Could not re-queue. Try again later.")
					}
				}
			case "\\a":
				m.autoRequeue = !m.autoRequeue
				var status string
				if m.autoRequeue {
					status = "enabled"
				} else {
					status = "disabled"
				}
				m.messages = append(m.messages, fmt.Sprintf("Auto-requeue %s. Send '\\h' for help.", status))
			case "\\q":
				switch m.chatState {
				case StateChatMatched:
					if err := m.user.LeaveChat(); err == nil {
						m.chatState = StateChatDisconnected
						m.messages = append(m.messages, "You have left the chat. Send '\\r' to requeue or press 'ctrl+c' to exit.")
					} else {
						m.messages = append(m.messages, "Error: Could not leave chat. Try again later.")
					}
				case StateChatQueued:
					if err := globalMatchmaker.Dequeue(m.user); err == nil {
						m.chatState = StateChatDisconnected
						m.messages = append(m.messages, "You have left the queue. Send '\\r' to requeue or press 'ctrl+c' to exit.")
					} else {
						m.messages = append(m.messages, "Error: Could not leave queue. Try again later.")
					}
				}
			default:
				if m.chatState == StateChatMatched {
					chatMsg := ChatMsg{
						Type:    ChatMsgTypeMessage,
						Content: strings.TrimSpace(m.textarea.Value()),
					}
					// Send message to other user (non-blocking)
					if err := m.user.SendMessage(chatMsg); err == nil {
						// Message sent successfully, add to our view
						m.messages = append(m.messages, m.senderStyle.Render("You: ")+m.textarea.Value())
					} else {
						// Channel is full or closed, show error
						m.messages = append(m.messages, "Error: Could not send message")
					}
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

	var view string

	switch m.uiState {
	case StateUIHelp:
		view = helpView(m)
	case StateUIChat:
		if m.chatState == StateChatMatched {
			m.textarea.Placeholder = "Type your message..."
		} else {
			m.textarea.Placeholder = "Waiting for match..." + m.splashSpinner.View()
		}
		view = fmt.Sprintf("%s%s%s", m.viewport.View(), gap, m.textarea.View())
	}

	return view
}

// splashView renders the splash screen animation.
func splashView(m model) string {
	var spinnerText string
	if m.splashTextIndex >= len(splashMessage) {
		spinnerText = m.splashSpinner.View()
	}
	return m.renderer.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		fmt.Sprintf("%s %s", m.splashText, m.splashStyle.Render(spinnerText)),
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
