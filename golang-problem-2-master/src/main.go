package main

import (
	"encoding/json"
	"errors"
	"io"
	"os"
)

const (
	KEYBOARD        = "keyboard"
	INLINE_KEYBOARD = "inline_keyboard"
)

type KeyboardButton struct {
	Text            string `json:"text"`
	RequestContact  bool   `json:"request_contact"`
	RequestLocation bool   `json:"request_location"`
}

type InlineKeyboardButton struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data"`
	Url          string `json:"url"`
}

type ReplyMarkup struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
	Keyboard       [][]KeyboardButton       `json:"keyboard"`
	ResizeKeyboard bool                     `json:"resize_keyboard"`
	OnTimeKeyboard bool                     `json:"one_time_keyboard"`
	Selective      bool                     `json:"selective"`
}

type SendMessage struct {
	ChatID      interface{} `json:"chat_id"`
	Text        string      `json:"text"`
	ParseMode   string      `json:"parse_mode"`
	ReplyMarkup interface{} `json:"reply_markup"`
}

type IKeyboard interface {
	InlineKeyboardButton | KeyboardButton
}

func DetectingKeyboardType[TypeKeyboard IKeyboard](replies interface{}, keyboard_type string) ([][]TypeKeyboard, error) {

	//	Function that makes an appropriate instance for our slice
	convToTypeKeyboard := func(reply interface{}, newKbType *TypeKeyboard, status int) error {
		var instance interface{}
		/*	Check status
			(1 -> KeyboardButton)
			(2 -> InlineKeyboardButton)
			(3 -> map[string]interface{})
		*/
		switch status {
		case 1:
			instance = KeyboardButton{Text: reply.(string)}
		case 2:
			instance = InlineKeyboardButton{Text: reply.(string)}
		case 3:
			instance = reply
		}

		//	Marshal the keyBtn which has to []byte
		data, err := json.Marshal(instance)
		if err != nil {
			return errors.New("can not marshal to []byte")
		}

		//	Unmarshal for type conversion from []byte to TypeKeyboard
		err = json.Unmarshal(data, &newKbType)
		if err != nil {
			return errors.New("can not unmarshal to TypeKeyboard")
		}
		return nil
	}

	//	Declare variable of InlineKeyboardButton or KeyboardButton for ReplyMarkup
	var kbButtons [][]TypeKeyboard
	//	Type Conversion (interface -> []interface)
	arrayReply := replies.([]interface{})
	//	Initialize a new slice for assigning []TypeKeyboard to each of it
	kbButtons = make([][]TypeKeyboard, len(arrayReply))
	for first_idx, items := range arrayReply {

		//	Type Conversion (interface -> []interface)
		arrayItems := items.([]interface{})

		//	Initialize a new slice for assigning TypeKeyboard to each of it
		kbButtons[first_idx] = make([]TypeKeyboard, len(arrayItems))
		for second_idx, item := range arrayItems {
			var newItem TypeKeyboard

			//	Assigning new instance of TypeKeyboard to each element of kbButtons
			//	Check type of item (string or map[string]interface)
			switch reply := item.(type) {
			case string:

				if keyboard_type == KEYBOARD {
					//	Create new instance of KeyboardButton and give it to kbButtons[first_idx][second_idx]
					convToTypeKeyboard(reply, &newItem, 1)
				} else {
					//	Create new instance of InlineKeyboardButton and give it to kbButtons[first_idx][second_idx]
					convToTypeKeyboard(reply, &newItem, 2)
				}

				kbButtons[first_idx][second_idx] = newItem

			case map[string]interface{}:
				//	Unmarshal reply (either InlineKeyboardButton or KeyboardButton) and give it to kbButtons[first_idx][second_idx]
				convToTypeKeyboard(reply, &newItem, 3)
				kbButtons[first_idx][second_idx] = newItem
			}
		}
	}
	return kbButtons, nil
}

func ReadSendMessageRequest(fileName string) (*SendMessage, error) {

	parseToReplyMarkup := func(reply map[string]interface{}) (
		*ReplyMarkup, error) {
		var err error
		//	Check Whether reply type is keyboard button or inline keyboard button
		if keyboard, exist := reply[KEYBOARD]; exist {
			reply[KEYBOARD], err = DetectingKeyboardType[KeyboardButton](keyboard, KEYBOARD)
			if err != nil {
				return nil, errors.New("can not get 'keyboard' in replies")
			}
		}
		if inlineKeyboard, exist := reply[INLINE_KEYBOARD]; exist {
			reply[INLINE_KEYBOARD], err = DetectingKeyboardType[InlineKeyboardButton](inlineKeyboard, INLINE_KEYBOARD)
			if err != nil {
				return nil, errors.New("can not get 'inline_keyboard' in replies")
			}
		}

		//	Marshal the replies which has map[string]interface{} to []byte
		data, err := json.Marshal(reply)
		if err != nil {
			return nil, errors.New("replies: can not marshal map[string]interface{} to []byte")
		}

		//	Unmarshal for type conversion from []byte to ReplyMarkup
		var replyMrkUp ReplyMarkup
		err = json.Unmarshal(data, &replyMrkUp)
		if err != nil {
			return nil, errors.New("can not unmarshal to ReplyMarkup")
		}

		return &replyMrkUp, nil

	}

	//	Open json file
	reader, err := os.Open(fileName)
	if err != nil {
		return nil, errors.New("can not open the given json file")
	}
	defer reader.Close()

	//	Read the given json file
	inputMsg, err := io.ReadAll(reader)
	if err != nil {
		return nil, errors.New("can not read data in json file")
	}

	//	Unmarshal the message to `SendMessage` struct (not correct json format)
	message := &SendMessage{}
	err = json.Unmarshal(inputMsg, message)
	if err != nil {
		return nil, errors.New("can not unmarshal given json")
	}

	//	Check existence of chat_id
	if message.ChatID == nil {
		return nil, errors.New("chat_id is empty")
	}

	//	Check text is not empty
	if message.Text == "" {
		return nil, errors.New("text is empty")
	}

	//	Check ReplyMarkup existence
	if message.ReplyMarkup != nil {
		//	Check type of replies whether it is string or map[string]interface{}
		switch replies := message.ReplyMarkup.(type) {

		//	Reply type is map[string]interface{}
		case map[string]interface{}:
			//	Parse what is in JSON files if the value of reply_markup is 2D array
			ptrReplyMarkup, err := parseToReplyMarkup(replies)
			if err != nil {
				return nil, err
			}
			message.ReplyMarkup = *ptrReplyMarkup

		//	Reply type is string
		case string:
			//	Define a variable with type map[string]interface{} to parse from string to it
			var replyMap map[string]interface{}
			json.Unmarshal([]byte(replies), &replyMap)

			//	Parse what is in JSON files if the value of reply_markup is string
			ptrReplyMarkup, err := parseToReplyMarkup(replyMap)
			if err != nil {
				return nil, err
			}
			message.ReplyMarkup = *ptrReplyMarkup
		}

		// if inlineKeyBtn := replyMrkUp.InlineKeyboard; inlineKeyBtn != nil {
		// 	//	Marshal the inlineKetBtn which has map[string]interface{} to []byte
		// 	data, err := json.Marshal(inlineKeyBtn)
		// 	if err != nil {
		// 		return nil, errors.New("can not marshal map[string]interface{} to []byte")
		// 	}

		// 	//	Unmarshal for type conversion from []byte to InlineKeyboardButton
		// 	var inlineKeyboardBtn [][]InlineKeyboardButton
		// 	err = json.Unmarshal(data, &inlineKeyboardBtn)
		// 	if err != nil {
		// 		return nil, errors.New("can not unmarshal to InlineKeyboardButton")
		// 	}

		// 	replyMrkUp.InlineKeyboard = inlineKeyboardBtn
		// }
		// if keyBtn := replyMrkUp.Keyboard; keyBtn != nil {
		// 	//	Marshal the keyBtn which has map[string]interface{} to []byte
		// 	data, err := json.Marshal(keyBtn)
		// 	if err != nil {
		// 		return nil, errors.New("can not marshal map[string]interface{} to []byte")
		// 	}

		// 	//	Unmarshal for type conversion from []byte to KeyboardButton
		// 	var keyboardBtn [][]KeyboardButton
		// 	err = json.Unmarshal(data, &keyboardBtn)
		// 	if err != nil {
		// 		return nil, errors.New("can not unmarshal to KeyboardButton")
		// 	}

		// 	replyMrkUp.Keyboard = keyBtn
		// }

	}
	return message, nil

}
