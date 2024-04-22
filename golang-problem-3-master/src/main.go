package main

import (
	"errors"
	"regexp"
	"time"
)

type Bale interface {
	AddUser(username string, isBot bool) (int, error)
	AddChat(chatname string, isGroup bool, creator int, admins []int) (int, error)
	SendMessage(userId, chatId int, text string) (int, error)
	SendLike(userId, messageId int) error
	GetNumberOfLikes(messageId int) (int, error)
	SetChatAdmin(chatId, userId int) error
	GetLastMessage(chatId int) (string, int, error)
	GetLastUserMessage(userId int) (string, int, error)
}

// User represents user entity
type User struct {
	Username string
	isBot    bool
}

// Chat represents chat entity
type Chat struct {
	ChatName string
	isGroup  bool
	Creator  int
}

// Message represents message entity
type Message struct {
	SenderId int
	ChatId   int
	Text     string
	SendAt   time.Time
}

// BaleImpl represents whole structure of bale
type BaleImpl struct {
	// User Section
	Users map[int]User
	// Check whether Usernames are unique or not
	Usernames       map[string]int
	LastAddedUserId int

	// Chat Section
	Chats           map[int]Chat
	Admins          map[int][]int
	LastAddedChatId int

	// Message Section
	Messages      map[int]Message
	MessageLikes  map[int][]int
	LastMessageId int
}

// NewBaleImpl creates new instance of Bale interface
func NewBaleImpl() Bale {
	return &BaleImpl{
		Users:        make(map[int]User),
		Usernames:    make(map[string]int),
		Chats:        make(map[int]Chat),
		Admins:       make(map[int][]int),
		Messages:     make(map[int]Message),
		MessageLikes: make(map[int][]int),
	}
}

// AddUser creates new user
func (b *BaleImpl) AddUser(username string, isBot bool) (int, error) {
	//	Check if there is duplicate or invalid username
	if _, ok := b.Usernames[username]; ok || len(username) <= 3 ||
		!regexp.MustCompile(
			//	Check that username contains both digit and letter
			`^[a-zA-Z0-9]*[a-z][a-zA-Z0-9]*[0-9][a-zA-Z0-9]*$|^[a-zA-Z0-9]*[0-9][a-zA-Z0-9]*[a-zA-Z][a-zA-Z0-9]*$`).MatchString(username) {
		return 0, errors.New("invalid username")
	}

	//	User created successfully
	newId := b.LastAddedUserId + 1
	b.Users[newId] = User{username, isBot}
	b.Usernames[username] = newId
	b.LastAddedUserId = newId
	return newId, nil
}

// AddChat creates new chat
func (b *BaleImpl) AddChat(chatname string, isGroup bool, creator int, admins []int) (int, error) {
	//	Check existence of creator
	creatorUser, exist := b.Users[creator]
	if !exist {
		return 0, errors.New("user not found")
	}

	//	Check that if all admins are valid or not
	for _, admin_id := range admins {
		if _, exist := b.Users[admin_id]; !exist {
			return 0, errors.New("admins are not valid")
		}
	}

	//	Check that user is not a bot
	if creatorUser.isBot {
		return 0, errors.New("could not create chat")
	}

	//	Chat created successfully
	newChatId := b.LastAddedChatId + 1
	b.Chats[newChatId] = Chat{chatname, isGroup, creator}
	b.Admins[newChatId] = admins
	b.LastAddedChatId = newChatId
	return newChatId, nil
}

// SendMessage creates new message
func (b *BaleImpl) SendMessage(userId, chatId int, text string) (int, error) {
	//	Check sender existence
	if _, exist := b.Users[userId]; !exist {
		return 0, errors.New("sender not found")
	}

	//	Check chat existence
	chat, exist := b.Chats[chatId]
	if !exist {
		return 0, errors.New("chat not found")
	}

	//	Check if the sender is not admin of the chat with type channel
	if !chat.isGroup {
		isAdmin := false
		for _, adminId := range b.Admins[chatId] {
			if adminId == userId {
				isAdmin = true
				break
			}
		}
		if !isAdmin {
			return 0, errors.New("user could not send message")
		}
	}

	//	Message sent successfully
	newMsgId := b.LastMessageId + 1
	b.Messages[newMsgId] = Message{userId, chatId, text, time.Now()}
	b.LastMessageId = newMsgId
	return newMsgId, nil

}

// SendLike like a message by a user
func (b *BaleImpl) SendLike(userId, messageId int) error {
	//	Check user existence
	if _, ok := b.Users[userId]; !ok {
		return errors.New("user not found")
	}

	//	Check message existence
	if _, ok := b.Messages[messageId]; !ok {
		return errors.New("message not found")
	}

	//	Check a message has already been liked by the user
	for _, userLikedId := range b.MessageLikes[messageId] {
		if userId == userLikedId {
			return errors.New("this user has liked this message before")
		}
	}

	//	Message has liked successfully
	b.MessageLikes[messageId] = append(b.MessageLikes[messageId], userId)
	return nil
}

// GetNumberOfLikes retrieves number of likes of a message
func (b *BaleImpl) GetNumberOfLikes(messageId int) (int, error) {
	//	Check message existence
	if _, ok := b.Messages[messageId]; !ok {
		return 0, errors.New("message not found")
	}

	//	Number of likes retrieved successfully
	return len(b.MessageLikes[messageId]), nil
}

// SetChatAdmin add a user to the list of admins of a chat
func (b *BaleImpl) SetChatAdmin(chatId, userId int) error {
	//	Check user existence
	if _, ok := b.Users[userId]; !ok {
		return errors.New("user not found")
	}

	//	Check chat existence
	if _, ok := b.Chats[chatId]; !ok {
		return errors.New("chat not found")
	}

	//	Check user has already been admin before
	for _, usr_id := range b.Admins[chatId] {
		if usr_id == userId {
			return errors.New("user is already admin")
		}
	}

	//	User has been added to admins list successfully
	b.Admins[chatId] = append(b.Admins[chatId], userId)
	return nil
}

// GetLastMessage retrieves last message of a chat
func (b *BaleImpl) GetLastMessage(chatId int) (string, int, error) {
	//	Check chat existence
	if _, exist := b.Chats[chatId]; !exist {
		return "", 0, errors.New("chat not found")
	}

	//	Find last message in the chat
	lastestMsgId := 0
	lastMsgTime := time.Time{}
	for idx, msg := range b.Messages {
		if msg.ChatId == chatId {
			if msg.SendAt.After(lastMsgTime) {
				lastMsgTime = msg.SendAt
				lastestMsgId = idx
			}
		}

	}

	//	Check for existence of the last message of the chat
	if lastestMsgId == 0 {
		return "", 0, errors.New("last message of the chat not found")
	}

	//	Retrieve last message of the chat with given chat id successfully
	return b.Messages[lastestMsgId].Text, lastestMsgId, nil
}

// GetLastUserMessage retrieves last message of a user
func (b *BaleImpl) GetLastUserMessage(userId int) (string, int, error) {
	//	Check user existence
	if _, exist := b.Users[userId]; !exist {
		return "", 0, errors.New("user not found")
	}

	//	Find last message of the user
	lastestMsgId := 0
	lastMsgTime := time.Time{}
	for idx, msg := range b.Messages {
		if msg.SenderId == userId {
			if msg.SendAt.After(lastMsgTime) {
				lastMsgTime = msg.SendAt
				lastestMsgId = idx
			}
		}

	}

	//	Check for existence of the last message of the user
	if lastestMsgId == 0 {
		return "", 0, errors.New("last message of the chat not found")
	}

	//	Retrieve last message of the user with given user id successfully
	return b.Messages[lastestMsgId].Text, lastestMsgId, nil
}
