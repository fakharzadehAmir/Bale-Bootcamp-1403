package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSendMessageSample1(t *testing.T) {
	msg, err := ReadSendMessageRequest("input_sample2.json")
	assert.Nil(t, err)
	assert.NotNil(t, msg)
	assert.EqualValues(t, "456", msg.ChatID)
	assert.IsType(t, ReplyMarkup{}, msg.ReplyMarkup)
	assert.IsType(t, msg.ReplyMarkup.(ReplyMarkup).InlineKeyboard, [][]InlineKeyboardButton{})
	assert.Equal(t, 2, len(msg.ReplyMarkup.(ReplyMarkup).InlineKeyboard))
	assert.Equal(t, 3, len(msg.ReplyMarkup.(ReplyMarkup).InlineKeyboard[0]))
	assert.Equal(t, "bale", msg.ReplyMarkup.(ReplyMarkup).InlineKeyboard[0][1].Text)
	assert.Equal(t, "HTML", msg.ParseMode)
}

func TestSendMessageSample2(t *testing.T) {
	msg, err := ReadSendMessageRequest("input_sample1.json")
	assert.NotNil(t, err)
	assert.Nil(t, msg)
	assert.Equal(t, "chat_id is empty", err.Error())
}

func TestSendMessageSample3(t *testing.T) {
	msg, err := ReadSendMessageRequest("input_sample3.json")
	assert.Nil(t, err)
	assert.NotNil(t, msg)
	assert.EqualValues(t, "1", msg.ChatID)
	assert.IsType(t, ReplyMarkup{}, msg.ReplyMarkup)
	assert.IsType(t, msg.ReplyMarkup.(ReplyMarkup).InlineKeyboard, [][]InlineKeyboardButton{})
	assert.Equal(t, 4, len(msg.ReplyMarkup.(ReplyMarkup).InlineKeyboard))
	assert.Equal(t, 4, len(msg.ReplyMarkup.(ReplyMarkup).InlineKeyboard[0]))
	assert.Equal(t, "Option 8", msg.ReplyMarkup.(ReplyMarkup).InlineKeyboard[0][2].Text)
	assert.Equal(t, "/check_7", msg.ReplyMarkup.(ReplyMarkup).InlineKeyboard[2][1].CallbackData)
	assert.Equal(t, "https://example.com/10", msg.ReplyMarkup.(ReplyMarkup).InlineKeyboard[3][0].Url)
	assert.Equal(t, "Markdown", msg.ParseMode)
}

func TestSendMessageSample4(t *testing.T) {
	msg, err := ReadSendMessageRequest("input_sample4.json")
	assert.Nil(t, err)
	assert.NotNil(t, msg)
	assert.EqualValues(t, "2", msg.ChatID)
	assert.IsType(t, ReplyMarkup{}, msg.ReplyMarkup)
	assert.IsType(t, msg.ReplyMarkup.(ReplyMarkup).Keyboard, [][]KeyboardButton{})
	assert.Equal(t, 2, len(msg.ReplyMarkup.(ReplyMarkup).Keyboard))
	assert.Equal(t, 6, len(msg.ReplyMarkup.(ReplyMarkup).Keyboard[0]))
	assert.Equal(t, "Option 4", msg.ReplyMarkup.(ReplyMarkup).Keyboard[0][3].Text)
	assert.Equal(t, true, msg.ReplyMarkup.(ReplyMarkup).Keyboard[0][0].RequestContact)
	assert.Equal(t, false, msg.ReplyMarkup.(ReplyMarkup).Keyboard[1][0].RequestLocation)
	assert.Equal(t, "HTML", msg.ParseMode)
}

func TestSendMessageSample5(t *testing.T) {
	msg, err := ReadSendMessageRequest("input_sample5.json")
	assert.NotNil(t, err)
	assert.Nil(t, msg)
	assert.Equal(t, "text is empty", err.Error())

}

func TestSendMessageSample6(t *testing.T) {
	msg, err := ReadSendMessageRequest("input_sample6.json")
	assert.Nil(t, err)
	assert.NotNil(t, msg)
	assert.EqualValues(t, "6", msg.ChatID)
	assert.IsType(t, ReplyMarkup{}, msg.ReplyMarkup)
	assert.IsType(t, msg.ReplyMarkup.(ReplyMarkup).Keyboard, [][]KeyboardButton{})
	assert.Equal(t, 2, len(msg.ReplyMarkup.(ReplyMarkup).Keyboard))
	assert.Equal(t, 6, len(msg.ReplyMarkup.(ReplyMarkup).Keyboard[0]))
	assert.Equal(t, "Option 4", msg.ReplyMarkup.(ReplyMarkup).Keyboard[0][3].Text)
	assert.Equal(t, false, msg.ReplyMarkup.(ReplyMarkup).Keyboard[0][0].RequestContact)
	assert.Equal(t, false, msg.ReplyMarkup.(ReplyMarkup).Keyboard[1][0].RequestLocation)
	assert.Equal(t, "", msg.ParseMode)
}

func TestSendMessageSample7(t *testing.T) {
	msg, err := ReadSendMessageRequest("input_sample7.json")
	assert.Nil(t, err)
	assert.NotNil(t, msg)
	assert.EqualValues(t, "6", msg.ChatID)
	assert.IsType(t, ReplyMarkup{}, msg.ReplyMarkup)
	assert.IsType(t, msg.ReplyMarkup.(ReplyMarkup).Keyboard, [][]KeyboardButton{})
	assert.Equal(t, 2, len(msg.ReplyMarkup.(ReplyMarkup).InlineKeyboard))
	assert.Equal(t, 6, len(msg.ReplyMarkup.(ReplyMarkup).InlineKeyboard[0]))
	assert.Equal(t, "Option 4", msg.ReplyMarkup.(ReplyMarkup).InlineKeyboard[0][3].Text)
	assert.Equal(t, "Option 10", msg.ReplyMarkup.(ReplyMarkup).InlineKeyboard[1][0].Text)
	assert.Equal(t, "", msg.ReplyMarkup.(ReplyMarkup).InlineKeyboard[1][0].Url)
	assert.Equal(t, "", msg.ParseMode)
}

func TestSendMessageSample8(t *testing.T) {
	msg, err := ReadSendMessageRequest("input_sample8.json")
	assert.Nil(t, err)
	assert.NotNil(t, msg)
	assert.EqualValues(t, "6", msg.ChatID)
	assert.IsType(t, ReplyMarkup{}, msg.ReplyMarkup)
	assert.IsType(t, msg.ReplyMarkup.(ReplyMarkup).Keyboard, [][]KeyboardButton{})
	assert.Equal(t, 2, len(msg.ReplyMarkup.(ReplyMarkup).InlineKeyboard))
	assert.Equal(t, 6, len(msg.ReplyMarkup.(ReplyMarkup).InlineKeyboard[0]))
	assert.Equal(t, "Option 4", msg.ReplyMarkup.(ReplyMarkup).InlineKeyboard[0][3].Text)
	assert.Equal(t, "Option 10", msg.ReplyMarkup.(ReplyMarkup).InlineKeyboard[1][0].Text)
	assert.Equal(t, "", msg.ReplyMarkup.(ReplyMarkup).InlineKeyboard[1][0].Url)
	assert.Equal(t, "", msg.ParseMode)
}

func TestSendMessageSample9(t *testing.T) {
	msg, err := ReadSendMessageRequest("input_sample9.json")
	assert.Nil(t, err)
	assert.NotNil(t, msg)
	assert.EqualValues(t, "10", msg.ChatID)
	assert.IsType(t, ReplyMarkup{}, msg.ReplyMarkup)
	assert.IsType(t, msg.ReplyMarkup.(ReplyMarkup).Keyboard, [][]KeyboardButton{})
	assert.Equal(t, 2, len(msg.ReplyMarkup.(ReplyMarkup).Keyboard))
	assert.Equal(t, 6, len(msg.ReplyMarkup.(ReplyMarkup).Keyboard[0]))
	assert.Equal(t, "Option 41", msg.ReplyMarkup.(ReplyMarkup).Keyboard[0][3].Text)
	assert.Equal(t, false, msg.ReplyMarkup.(ReplyMarkup).Keyboard[0][0].RequestContact)
	assert.Equal(t, false, msg.ReplyMarkup.(ReplyMarkup).Keyboard[1][0].RequestLocation)
	assert.Equal(t, "", msg.ParseMode)
}

func TestSendMessageSample10(t *testing.T) {
	msg, err := ReadSendMessageRequest("input_sample10.json")
	assert.Nil(t, err)
	assert.NotNil(t, msg)
	assert.EqualValues(t, "2", msg.ChatID)
	assert.IsType(t, ReplyMarkup{}, msg.ReplyMarkup)
	assert.IsType(t, msg.ReplyMarkup.(ReplyMarkup).Keyboard, [][]KeyboardButton{})
	assert.Equal(t, 2, len(msg.ReplyMarkup.(ReplyMarkup).Keyboard))
	assert.Equal(t, 6, len(msg.ReplyMarkup.(ReplyMarkup).Keyboard[0]))
	assert.Equal(t, "Option 4", msg.ReplyMarkup.(ReplyMarkup).Keyboard[0][3].Text)
	assert.Equal(t, true, msg.ReplyMarkup.(ReplyMarkup).Keyboard[0][0].RequestContact)
	assert.Equal(t, false, msg.ReplyMarkup.(ReplyMarkup).Keyboard[1][0].RequestLocation)
	assert.Equal(t, "HTML", msg.ParseMode)
}

func TestSendMessageSample11(t *testing.T) {
	msg, err := ReadSendMessageRequest("input_sample11.json")
	assert.NotNil(t, err)
	assert.Nil(t, msg)
	assert.Equal(t, "can not unmarshal given json", err.Error())
}

func TestSendMessageSample12(t *testing.T) {
	msg, err := ReadSendMessageRequest("input_sample12.json")
	assert.NotNil(t, err)
	assert.Nil(t, msg)
	assert.Equal(t, "can not open the given json file", err.Error())
}

func TestSendMessageSample13(t *testing.T) {
	msg, err := ReadSendMessageRequest("input_sample13.json")
	assert.Nil(t, err)
	assert.NotNil(t, msg)
	assert.EqualValues(t, "890", msg.ChatID)
	assert.IsType(t, ReplyMarkup{}, msg.ReplyMarkup)
	assert.IsType(t, msg.ReplyMarkup.(ReplyMarkup).InlineKeyboard, [][]InlineKeyboardButton{})
	assert.Equal(t, 4, len(msg.ReplyMarkup.(ReplyMarkup).InlineKeyboard))
	assert.Equal(t, 4, len(msg.ReplyMarkup.(ReplyMarkup).InlineKeyboard[0]))
	assert.Equal(t, "Option 8", msg.ReplyMarkup.(ReplyMarkup).InlineKeyboard[0][2].Text)
	assert.Equal(t, "/check_7", msg.ReplyMarkup.(ReplyMarkup).InlineKeyboard[2][1].CallbackData)
	assert.Equal(t, "https://example.com/10", msg.ReplyMarkup.(ReplyMarkup).InlineKeyboard[3][0].Url)
	assert.Equal(t, "Markdown", msg.ParseMode)

}
