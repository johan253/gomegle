package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/proto"
)

// User represents a user in the matchmaker system
type User struct {
	pubKey  string                // Public key of the user
	pubsub  *redis.PubSub         // Redis PubSub instance for the user
	receive <-chan *redis.Message // Channel to receive messages
	send    string                // Channel to send messages
}

// ListenForMessages starts listening for messages on the user's receive channel
func (u *User) ListenForMessages() tea.Cmd {
	return func() tea.Msg {
		content, ok := <-u.receive // Blocking call to wait for a message
		if !ok {
			return tea.Quit // If the channel is closed, quit the program
		}
		msg := &ChatMsg{}
		if err := proto.Unmarshal([]byte(content.Payload), msg); err != nil {
			return tea.Quit // If there's an error, quit the program
		}
		return chatMsgReceived(msg)
	}
}

// SendMessage sends a message to the user's match channel. If the
// send channel (string identifier) is empty, this function does nothing.
func (u *User) SendMessage(msg *ChatMsg) error {
	if u.send == "" {
		return nil // If the send channel is not set, do nothing
	}
	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	return rdb.Publish(ctx, "user:"+u.send, data).Err()
}

func (u *User) LeaveChat() error {
	leaveMsg := &ChatMsg{
		Type:    ChatMsgTypeLeave,
		Content: "Stranger has left the chat",
	}
	err := u.SendMessage(leaveMsg)
	if err != nil {
		return err
	}
	u.send = ""
	return nil
}
