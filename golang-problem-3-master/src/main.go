package main

type Bale interface {
	AddUser(username string, isBot bool) (int, error)
	AddChat(chatname string, isGroup bool, creator int, admins []int) (int, error)
	SendMessage(userId, chatId int, text string) (int, error)
	SendLike(userId, messageId int) error
	GetNumberOfLikes(messageId int) (int, error)
	SetChatAdmin(chatId, userId int) error
	GetLastMessage(chatId int) (string, int, error)
	GetLastUserMessage(userId int) (string ,int, error)
}



type BaleImpl struct {
	// TODO : Implement Bale
}

func NewBaleImpl() *BaleImpl {
	panic("Implement me!")
}