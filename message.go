package main

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
