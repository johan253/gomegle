package main

import (
	"sync"

	tea "github.com/charmbracelet/bubbletea"
)

type ChatMsgType int

const (
	ChatMsgTypeMessage ChatMsgType = iota
	ChatMsgTypeJoin
	ChatMsgTypeLeave
	ChatMsgTypeError
)

type ChatMsg struct {
	Type    ChatMsgType // Type of message (e.g., message, join, leave, error)
	Content string      // Content of the message
}

// chatMsgReceived is a custom type to handle received chat messages in tea
type chatMsgReceived ChatMsg

type User struct {
	receive chan ChatMsg // Channel to receive messages
	send    chan ChatMsg // Channel to send messages
}

func (u *User) ListenForMessages() tea.Cmd {
	return func() tea.Msg {
		msg := <-u.receive // Blocking call to wait for a message
		return chatMsgReceived(msg)
	}
}

type Matchmaker struct {
	queue []*User       // Queue of users waiting to be matched
	mu    sync.Mutex    // Mutex to protect access to the queue
	added chan struct{} // Channel to signal a new user has been added
}

func NewMatchMaker() *Matchmaker {
	m := &Matchmaker{
		queue: make([]*User, 0),
		mu:    sync.Mutex{},
		added: make(chan struct{}, 1000), // Buffered channel to avoid blocking
	}
	go m.matchUsers()
	return m
}

func (m *Matchmaker) matchUsers() {
	for {
		m.mu.Lock()
		for len(m.queue) < 2 {
			m.mu.Unlock() // Unlock while waiting for a new user
			<-m.added     // Wait for a new user to be added
			m.mu.Lock()   // Lock again to access the queue
		}
		// Match the first two users in the queue
		u1 := m.queue[0]
		u2 := m.queue[1]
		m.queue = m.queue[2:] // Remove the matched users from the queue
		m.mu.Unlock()

		// Set up the channels for the matched users
		u1.send = u2.receive // Set up the send channel for user 1
		u2.send = u1.receive // Set up the send channel for user 2

		// Notify both users that they have been matched
		joinMsg := ChatMsg{
			Type:    ChatMsgTypeJoin,
			Content: "You have been matched with another user!",
		}
		u1.receive <- joinMsg
		u2.receive <- joinMsg
	}
}

func (m *Matchmaker) Enqueue(u *User) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Add the user to the queue (store pointer to maintain channel references)
	m.queue = append(m.queue, u)

	// Signal that a new user has been added
	select {
	case m.added <- struct{}{}:
	default: // Don't block if channel is full
	}
}

// Dequeue removes a user from the queue and closes their send channel. If the user is
// not found, this function is a no-op.
func (m *Matchmaker) Dequeue(u *User) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Find the user in the queue and remove them
	for i, user := range m.queue {
		if user == u {
			// Close the send channel to signal that the user has left
			m.queue = append(m.queue[:i], m.queue[i+1:]...) // Remove user from queue
			return
		}
	}
}
