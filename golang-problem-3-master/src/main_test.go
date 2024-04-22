package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// AddUser Section
// Test 1: len(username) lower than 3
func TestSample_AddUser1(t *testing.T) {
	b := NewBaleImpl()
	_, err := b.AddUser("a", false)
	assert.NotNil(t, err)
	assert.Equal(t, "invalid username", err.Error())
}

// Test 2: invalid username because of not having digit or letter
func TestSample_AddUser2(t *testing.T) {
	b := NewBaleImpl()

	//	just letter
	_, err := b.AddUser("abc", false)
	assert.NotNil(t, err)
	assert.Equal(t, "invalid username", err.Error())

	//	just digit
	_, err = b.AddUser("123", false)
	assert.NotNil(t, err)
	assert.Equal(t, "invalid username", err.Error())
}

// Test 3: check username existence
func TestSample_AddUser3(t *testing.T) {
	b := NewBaleImpl()
	uId, err := b.AddUser("abc123", false)
	assert.Equal(t, 1, uId)
	assert.Nil(t, err)

	//	creating existed username
	_, err = b.AddUser("abc123", false)
	assert.NotNil(t, err)
	assert.Equal(t, "invalid username", err.Error())
}

// Test 4: user has been created successfully
func TestSample_AddUser4(t *testing.T) {
	b := NewBaleImpl()
	uId_1, err := b.AddUser("abc123", false)
	assert.Nil(t, err)
	assert.Equal(t, 1, uId_1)
	uId_2, err := b.AddUser("abc456", true)
	assert.Nil(t, err)
	assert.Equal(t, 2, uId_2)
}

// AddChat Section
// Test 5: user not found, invalid admins
func TestSample_AddChat1(t *testing.T) {
	b := NewBaleImpl()
	uId_1, _ := b.AddUser("abc123", false)
	uId_2, _ := b.AddUser("abc456", false)
	uId_3, _ := b.AddUser("abc789", false)

	//	user not found
	_, err := b.AddChat("chat 1", true, uId_3+1, []int{uId_3})
	assert.NotNil(t, err)
	assert.Equal(t, "user not found", err.Error())

	//	invalid admins
	_, err = b.AddChat("chat 1", true, uId_1, []int{uId_1, uId_2, uId_3 + 1})
	assert.NotNil(t, err)
	assert.Equal(t, "admins are not valid", err.Error())

}

// Test 6: chat creation with failure because of user mode
func TestSample_AddChat2(t *testing.T) {
	b := NewBaleImpl()
	uId, _ := b.AddUser("abc123", true)
	_, err := b.AddChat("first_chat", true, uId, []int{uId})
	assert.NotNil(t, err)
	assert.Equal(t, "could not create chat", err.Error())
}

// Test 7: chat has been created successfully
func TestSample_AddChat3(t *testing.T) {
	b := NewBaleImpl()
	uId_1, _ := b.AddUser("abc123", false)
	uId_2, _ := b.AddUser("abc456", false)
	uId_3, _ := b.AddUser("abc789", false)

	//	chat with 1 admin
	chat_id, err := b.AddChat("chat 1 admin", true, uId_1, []int{uId_1})
	assert.Nil(t, err)
	assert.Equal(t, 1, chat_id)

	//	chat with more than 1 admibn
	chat_id, err = b.AddChat("chat more than 1 admin", true, uId_1, []int{uId_1, uId_2, uId_3})
	assert.Nil(t, err)
	assert.Equal(t, 2, chat_id)
}

// SendMessage Section
//
//	Test 8: messages and user not found
func TestSample_SendMessage1(t *testing.T) {
	b := NewBaleImpl()
	uId_1, _ := b.AddUser("abc123", false)
	chat_id, _ := b.AddChat("chat_with_more_admin", true, uId_1, []int{uId_1})

	//	user not found
	_, err := b.SendMessage(uId_1+1, chat_id, "Salam in payam e admin hast")
	assert.NotNil(t, err)
	assert.Equal(t, "sender not found", err.Error())

	//	chat not found
	_, err = b.SendMessage(uId_1, chat_id+1, "Salam in payam e admin hast")
	assert.NotNil(t, err)
	assert.Equal(t, "chat not found", err.Error())
}

// Test 9: sending a message in a chat (type channel) through a non-admin user
func TestSample_SendMessage2(t *testing.T) {
	b := NewBaleImpl()
	uId_1, _ := b.AddUser("abc123", false)
	uId_2, _ := b.AddUser("abc456", false)
	uId_3, _ := b.AddUser("abc789", false)
	chat_id, _ := b.AddChat("chat_with_more_admin", false, uId_1, []int{uId_1, uId_2})

	_, err := b.SendMessage(uId_3, chat_id, "Salam in payam e yek non-admin hast")
	assert.NotNil(t, err)
	assert.Equal(t, "user could not send message", err.Error())
}

// Test 10: sending message in a chat (type channel) through its admins
func TestSample_SendMessage3(t *testing.T) {
	b := NewBaleImpl()
	uId_1, _ := b.AddUser("abc123", false)
	uId_2, _ := b.AddUser("abc456", false)
	chat_id, _ := b.AddChat("chat_1", false, uId_1, []int{uId_1, uId_2})

	msgId_1, err := b.SendMessage(uId_1, chat_id, "Salam in payam e admin - 1 hast")
	assert.Nil(t, err)
	assert.Equal(t, 1, msgId_1)

	msgId_2, err := b.SendMessage(uId_2, chat_id, "Salam in payam e yek admin - 2 hast")
	assert.Nil(t, err)
	assert.Equal(t, 2, msgId_2)
}

// Test 11: sending message in a chat (type group) successfully
func TestSample_SendMessage4(t *testing.T) {
	b := NewBaleImpl()
	uId_1, _ := b.AddUser("abc123", false)
	uId_2, _ := b.AddUser("abc456", false)
	chat_id, _ := b.AddChat("chat_1", true, uId_1, []int{uId_1})

	msgId_1, err := b.SendMessage(uId_1, chat_id, "Salam in payam e admin hast")
	assert.Nil(t, err)
	assert.Equal(t, 1, msgId_1)

	msgId_2, err := b.SendMessage(uId_2, chat_id, "Salam in payam e yek non-admin hast")
	assert.Nil(t, err)
	assert.Equal(t, 2, msgId_2)
}

// SendLike Section
// Test 12: user has liked a message successfully
func TestSample_SendLike1(t *testing.T) {
	b := NewBaleImpl()
	uId_1, _ := b.AddUser("abc123", false)
	uId_2, _ := b.AddUser("abc456", false)
	chat_id, _ := b.AddChat("chat_with_more_admin", true, uId_1, []int{uId_1})
	msgId, _ := b.SendMessage(uId_1, chat_id, "Salam in payam e admin hast")
	err := b.SendLike(uId_2, msgId)
	assert.Nil(t, err)
}

// Test 13: user has liked a message again
func TestSample_SendLike2(t *testing.T) {
	b := NewBaleImpl()
	uId_1, _ := b.AddUser("abc123", false)
	chat_id, _ := b.AddChat("chat_with_more_admin", true, uId_1, []int{uId_1})
	msgId, _ := b.SendMessage(uId_1, chat_id, "Salam in payam e admin hast")

	//	User 1 user has liked a message
	_ = b.SendLike(uId_1, msgId)

	//	User 1 user has liked the message again
	err := b.SendLike(uId_1, msgId)
	assert.NotNil(t, err)
	assert.Equal(t, "this user has liked this message before", err.Error())
}

// Test 14: user and message not found
func TestSample_SendLike3(t *testing.T) {
	b := NewBaleImpl()
	uId_1, _ := b.AddUser("abc123", false)
	chat_id, _ := b.AddChat("chat_with_more_admin", true, uId_1, []int{uId_1})
	msgId, _ := b.SendMessage(uId_1, chat_id, "Salam in payam e admin hast")
	//	user not found
	err := b.SendLike(uId_1+1, msgId)
	assert.NotNil(t, err)
	assert.Equal(t, "user not found", err.Error())

	//	message not found
	err = b.SendLike(uId_1, msgId+1)
	assert.NotNil(t, err)
	assert.Equal(t, "message not found", err.Error())
}

// GetNumberOfLikes Section
//
//	Test 15: retrieve number of likes successfully
func TestSample_GetNumberOfLikes1(t *testing.T) {
	b := NewBaleImpl()
	uId_1, _ := b.AddUser("abc123", false)
	chat_id, _ := b.AddChat("chat_with_more_admin", true, uId_1, []int{uId_1})
	msgId, _ := b.SendMessage(uId_1, chat_id, "Salam in payam e admin hast")

	//	User 1 user has liked a message
	_ = b.SendLike(uId_1, msgId)
	numLike, _ := b.GetNumberOfLikes(msgId)
	assert.Equal(t, 1, numLike)

	//	User 1 user has liked the message again
	_ = b.SendLike(uId_1, msgId)
	numLike, _ = b.GetNumberOfLikes(msgId)
	assert.Equal(t, 1, numLike)

	//	More user has liked the message
	uId_2, _ := b.AddUser("abc456", false)
	uId_3, _ := b.AddUser("abc789", false)
	uId_4, _ := b.AddUser("abc000", false)
	_ = b.SendLike(uId_2, msgId)
	_ = b.SendLike(uId_3, msgId)
	_ = b.SendLike(uId_4, msgId)
	numLike, _ = b.GetNumberOfLikes(msgId)
	assert.Equal(t, 4, numLike)
}

// Test 16: message not found
func TestSample_GetNumberOfLikes2(t *testing.T) {
	b := NewBaleImpl()
	uId_1, _ := b.AddUser("abc123", false)
	chat_id, _ := b.AddChat("chat_with_more_admin", true, uId_1, []int{uId_1})
	msgId, _ := b.SendMessage(uId_1, chat_id, "Salam in payam e admin hast")

	_ = b.SendLike(uId_1, msgId)
	_, err := b.GetNumberOfLikes(msgId + 1)
	assert.NotNil(t, err)
	assert.Equal(t, "message not found", err.Error())
}

//	SetChatAdmin Section
//
// Test 17: user has been added to list of admins successfully
func TestSample_SetChatAdmin1(t *testing.T) {
	b := NewBaleImpl()

	//	Type of chat is group
	uId_1, _ := b.AddUser("abc123", false)
	chatId_1, _ := b.AddChat("chat1", true, uId_1, []int{uId_1})

	uId_2, _ := b.AddUser("abc456", false)
	err := b.SetChatAdmin(chatId_1, uId_2)
	assert.Nil(t, err)

	uId_3, _ := b.AddUser("abc789", true)
	err = b.SetChatAdmin(chatId_1, uId_3)
	assert.Nil(t, err)

	//	Type of chat is channel
	chatId_2, _ := b.AddChat("chat1", true, uId_1, []int{uId_1})

	err = b.SetChatAdmin(chatId_2, uId_2)
	assert.Nil(t, err)

	err = b.SetChatAdmin(chatId_2, uId_3)
	assert.Nil(t, err)
}

// Test 18: user and chat not found
func TestSample_SetChatAdmin2(t *testing.T) {
	b := NewBaleImpl()

	uId_1, _ := b.AddUser("abc123", false)
	chat_id, _ := b.AddChat("chat1", true, uId_1, []int{uId_1})

	uId_2, _ := b.AddUser("abc456", false)

	//	User not found
	err := b.SetChatAdmin(chat_id, uId_2+1)
	assert.NotNil(t, err)
	assert.Equal(t, "user not found", err.Error())

	//	Chat not found
	err = b.SetChatAdmin(chat_id+1, uId_2)
	assert.NotNil(t, err)
	assert.Equal(t, "chat not found", err.Error())
}

// Test 19: user has already been added to the list of admins before
func TestSample_SetChatAdmin3(t *testing.T) {
	b := NewBaleImpl()

	uId_1, _ := b.AddUser("abc123", false)
	chat_id, _ := b.AddChat("chat1", true, uId_1, []int{uId_1})

	//	User 1 added before
	err := b.SetChatAdmin(chat_id, uId_1)
	assert.NotNil(t, err)
	assert.Equal(t, "user is already admin", err.Error())

	//	Adding User 2 to the list of admins
	uId_2, _ := b.AddUser("abc456", false)

	err = b.SetChatAdmin(chat_id, uId_2)
	assert.Nil(t, err)

	//	User 2 added before
	err = b.SetChatAdmin(chat_id, uId_2)
	assert.NotNil(t, err)
	assert.Equal(t, "user is already admin", err.Error())

}

// GetLastMessage Section
// Test 20: chat not found
func TestSample_GetLastMessage1(t *testing.T) {
	b := NewBaleImpl()
	uId_1, _ := b.AddUser("abc123", false)
	chat_id, _ := b.AddChat("chat1", false, uId_1, []int{uId_1})
	_, _ = b.SendMessage(uId_1, chat_id, "Salam in payam e admin hast")

	_, _, err := b.GetLastMessage(chat_id + 1)
	assert.NotNil(t, err)
	assert.Equal(t, "chat not found", err.Error())
}

// Test 21: chat last message not found
func TestSample_GetLastMessage2(t *testing.T) {
	b := NewBaleImpl()
	uId_1, _ := b.AddUser("abc123", false)
	chat_id, _ := b.AddChat("chat1", false, uId_1, []int{uId_1})

	//	No message has been sent in the chat

	_, _, err := b.GetLastMessage(chat_id)
	assert.NotNil(t, err)
	assert.Equal(t, "last message of the chat not found", err.Error())
}

// Test 22: retrieve last message of chat successfully (channel senario)
func TestSample_GetLastMessage3(t *testing.T) {
	b := NewBaleImpl()

	uId_1, _ := b.AddUser("abc123", false)
	chat_id, _ := b.AddChat("chat1", false, uId_1, []int{uId_1})

	//	Message has been sent through an admin
	_, _ = b.SendMessage(uId_1, chat_id, "salam in channel hastesh va man adminam")
	text_msg, id_msg, err := b.GetLastMessage(chat_id)
	assert.Nil(t, err)
	assert.Equal(t, "salam in channel hastesh va man adminam", text_msg)
	assert.Equal(t, id_msg, 1)

	uId_2, _ := b.AddUser("abc456", true)

	//	Message has been sent through a non-admin
	_, _ = b.SendMessage(uId_2, chat_id, "admin nistam")

	//	Retrieved message is message of user 1 not user 2 (not an admin)
	text_msg, id_msg, err = b.GetLastMessage(chat_id)
	assert.Nil(t, err)

	assert.NotEqual(t, "admin nistam", text_msg)
	assert.NotEqual(t, id_msg, 2)

	assert.Equal(t, "salam in channel hastesh va man adminam", text_msg)
	assert.Equal(t, id_msg, 1)

	//	Add User 2 to admins of the chat and sending a message
	_ = b.SetChatAdmin(chat_id, uId_2)
	_, _ = b.SendMessage(uId_2, chat_id, "admin shodam")

	//	Check the new message for User 2 (new admin)
	text_msg, id_msg, err = b.GetLastMessage(chat_id)
	assert.Nil(t, err)
	assert.Equal(t, "admin shodam", text_msg)
	assert.Equal(t, id_msg, 2)

	//	Test more messages
	_, _ = b.SendMessage(uId_1, chat_id, "admin budam az aval")
	text_msg, id_msg, err = b.GetLastMessage(chat_id)
	assert.Nil(t, err)
	assert.Equal(t, "admin budam az aval", text_msg)
	assert.Equal(t, id_msg, 3)

	_, _ = b.SendMessage(uId_2, chat_id, "salam admine jadid oomad")
	text_msg, id_msg, err = b.GetLastMessage(chat_id)
	assert.Nil(t, err)
	assert.Equal(t, "salam admine jadid oomad", text_msg)
	assert.Equal(t, id_msg, 4)
}

// Test 23: Retrieved successfully generally
func TestSample_GetLastMessage4(t *testing.T) {
	b := NewBaleImpl()

	uId_1, _ := b.AddUser("abc123", false)
	chat_id, _ := b.AddChat("chat1", true, uId_1, []int{uId_1})

	//	Message has been sent through an admin
	_, _ = b.SendMessage(uId_1, chat_id, "salam 1")
	text_msg, id_msg, err := b.GetLastMessage(chat_id)
	assert.Nil(t, err)
	assert.Equal(t, "salam 1", text_msg)
	assert.Equal(t, id_msg, 1)

	//	Another admin
	uId_2, _ := b.AddUser("abc456", true)
	_, _ = b.SendMessage(uId_2, chat_id, "user 2")
	text_msg, id_msg, err = b.GetLastMessage(chat_id)
	assert.Nil(t, err)
	assert.Equal(t, "user 2", text_msg)
	assert.Equal(t, id_msg, 2)

	//	Test more messages
	_, _ = b.SendMessage(uId_1, chat_id, "admin 1")
	text_msg, id_msg, err = b.GetLastMessage(chat_id)
	assert.Nil(t, err)
	assert.Equal(t, "admin 1", text_msg)
	assert.Equal(t, id_msg, 3)

	_, _ = b.SendMessage(uId_2, chat_id, "user 2")
	text_msg, id_msg, err = b.GetLastMessage(chat_id)
	assert.Nil(t, err)
	assert.Equal(t, "user 2", text_msg)
	assert.Equal(t, id_msg, 4)
}

// GetLastUserMessage Section
// Test 24: user not found
func TestSample_GetLastUserMessage1(t *testing.T) {
	b := NewBaleImpl()
	uId_1, _ := b.AddUser("abc123", false)
	chat_id, _ := b.AddChat("chat1", false, uId_1, []int{uId_1})
	_, _ = b.SendMessage(uId_1, chat_id, "Salam in payam e admin hast")

	_, _, err := b.GetLastUserMessage(uId_1 + 1)
	assert.NotNil(t, err)
	assert.Equal(t, "user not found", err.Error())
}

// Test 25: user last message not found
func TestSample_GetLastUserMessage2(t *testing.T) {
	b := NewBaleImpl()
	uId_1, _ := b.AddUser("abc123", false)
	_, _ = b.AddChat("chat1", false, uId_1, []int{uId_1})

	//	No message has been sent in the chat by user 1

	_, _, err := b.GetLastUserMessage(uId_1)
	assert.NotNil(t, err)
	assert.Equal(t, "last message of the chat not found", err.Error())
}

// Test 26: retrieve last message of chat successfully (channel senario)
func TestSample_GetLastUserMessage3(t *testing.T) {
	b := NewBaleImpl()

	uId_1, _ := b.AddUser("abc123", false)
	uId_2, _ := b.AddUser("abc456", false)
	chatId_1, _ := b.AddChat("chat1", true, uId_1, []int{uId_1})
	chatId_2, _ := b.AddChat("chat2", true, uId_2, []int{uId_2})
	chatId_3, _ := b.AddChat("chat3", false, uId_2, []int{uId_2})
	chatId_4, _ := b.AddChat("chat4", false, uId_1, []int{uId_1})

	//	The user has sent a message to chat 1
	_, _ = b.SendMessage(uId_1, chatId_1, "salam 1")
	text_msg, id_msg, err := b.GetLastUserMessage(uId_1)
	assert.Nil(t, err)
	assert.Equal(t, "salam 1", text_msg)
	assert.Equal(t, id_msg, 1)

	//	Sending a message to a channel that user 2 is not in its admins
	_, _ = b.SendMessage(uId_1, chatId_3, "salam 3")
	text_msg, id_msg, err = b.GetLastUserMessage(uId_1)
	assert.Nil(t, err)
	assert.NotEqual(t, "salam 3", text_msg)
	assert.NotEqual(t, id_msg, 2)

	assert.Equal(t, "salam 1", text_msg)
	assert.Equal(t, id_msg, 1)

	//	Test more messages
	_, _ = b.SendMessage(uId_1, chatId_2, "salam 2")
	text_msg, id_msg, err = b.GetLastUserMessage(uId_1)
	assert.Nil(t, err)
	assert.Equal(t, "salam 2", text_msg)
	assert.Equal(t, id_msg, 2)

	_, _ = b.SendMessage(uId_1, chatId_4, "salam 4")
	text_msg, id_msg, err = b.GetLastUserMessage(uId_1)
	assert.Nil(t, err)
	assert.Equal(t, "salam 4", text_msg)
	assert.Equal(t, id_msg, 3)
}

// Test 27: Retrieved successfully generally
func TestSample_GetLastUserMessage4(t *testing.T) {
	b := NewBaleImpl()

	uId_1, _ := b.AddUser("abc123", false)
	uId_2, _ := b.AddUser("abc456", false)
	chatId_1, _ := b.AddChat("chat1", true, uId_1, []int{uId_1})
	chatId_2, _ := b.AddChat("chat2", true, uId_2, []int{uId_2})

	_, _ = b.SendMessage(uId_1, chatId_1, "salam 1")
	text_msg, id_msg, err := b.GetLastUserMessage(uId_1)
	assert.Nil(t, err)
	assert.Equal(t, "salam 1", text_msg)
	assert.Equal(t, id_msg, 1)

	_, _ = b.SendMessage(uId_2, chatId_1, "salam 2")
	text_msg, id_msg, err = b.GetLastUserMessage(uId_2)
	assert.Nil(t, err)
	assert.Equal(t, "salam 2", text_msg)
	assert.Equal(t, id_msg, 2)

	_, _ = b.SendMessage(uId_1, chatId_2, "salam 3")
	text_msg, id_msg, err = b.GetLastUserMessage(uId_1)
	assert.Nil(t, err)
	assert.Equal(t, "salam 3", text_msg)
	assert.Equal(t, id_msg, 3)

	_, _ = b.SendMessage(uId_2, chatId_2, "salam 4")
	text_msg, id_msg, err = b.GetLastUserMessage(uId_2)
	assert.Nil(t, err)
	assert.Equal(t, "salam 4", text_msg)
	assert.Equal(t, id_msg, 4)
}
