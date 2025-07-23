package main

import tea "github.com/charmbracelet/bubbletea"

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
