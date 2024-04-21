package main


type KeyboardButton struct {
	Text string `json:"text"`
	RequestContact bool `json:"request_contact"`
	RequestLocation bool `json:"request_location"`
}

type InlineKeyboardButton struct {
	Text string `json:"text"`
	CallbackData string `json:"callback_data"`
	Url string `json:"url"`
}

type ReplyMarkup struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
	Keyboard [][]KeyboardButton `json:"keyboard"`
	ResizeKeyboard bool `json:"resize_keyboard"`
	OnTimeKeyboard bool `json:"one_time_keyboard"`
	Selective bool `json:"selective"`
}


type SendMessage struct {
	ChatID           interface{} `json:"chat_id"`
	Text             string `json:"text"`
	ParseMode        string `json:"parse_mode"`
	ReplyMarkup      interface{} `json:"reply_markup"`
}





func ReadSendMessageRequest(fileName string) (*SendMessage, error) {
	panic("Implement me!")
}
