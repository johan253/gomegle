package main

// // Enum for chat message types
// type ChatMsgType int
const (
	ChatMsgTypeMessage ChatMsgType = ChatMsgType_MESSAGE // Regular message from a user
	ChatMsgTypeJoin    ChatMsgType = ChatMsgType_JOIN    // User has been matched with another user
	ChatMsgTypeLeave   ChatMsgType = ChatMsgType_LEAVE   // User has left the chat
	ChatMsgTypeError   ChatMsgType = ChatMsgType_ERROR   // Error message
)

//
// // ChatMsg represents a message in the chat system
// // It includes the type of message and its content.
// // This is used to communicate between user send and receive channels.
// type ChatMsg struct {
// 	Type    ChatMsgType // Type of message (e.g., message, join, leave, error)
// 	Content string      // Content of the message
// }
//
// // chatMsgReceived is a custom type to handle received chat messages in tea
// type chatMsgReceived ChatMsg
