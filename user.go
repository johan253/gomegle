package main

import tea "github.com/charmbracelet/bubbletea"

// User represents a user in the matchmaker system
type User struct {
	receive chan ChatMsg // Channel to receive messages
	send    chan ChatMsg // Channel to send messages
}

// ListenForMessages starts listening for messages on the user's receive channel
func (u *User) ListenForMessages() tea.Cmd {
	return func() tea.Msg {
		msg := <-u.receive // Blocking call to wait for a message
		return chatMsgReceived(msg)
	}
}
