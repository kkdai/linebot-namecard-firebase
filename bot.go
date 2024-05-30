package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/line/line-bot-sdk-go/v7/linebot"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
)

// Const variables of Prompts.
const ImagePrompt = "這是一張名片，你是一個名片秘書。請將以下資訊整理成 json 給我。如果看不出來的，幫我填寫 N/A， 只好 json 就好:  Name, Title, Address, Email, Phone, Company.   其中 Phone 的內容格式為 #886-0123-456-789,1234. 沒有分機就忽略 ,1234"

// replyText: Reply text message to LINE server.
func replyText(replyToken, text string) error {
	if _, err := bot.ReplyMessage(
		&messaging_api.ReplyMessageRequest{
			ReplyToken: replyToken,
			Messages: []messaging_api.MessageInterface{
				&messaging_api.TextMessage{
					Text: text,
				},
			},
		},
	); err != nil {
		return err
	}
	return nil
}

// callbackHandler: Handle callback from LINE server.
func callbackHandler(w http.ResponseWriter, r *http.Request) {
	card_prompt := os.Getenv("CARD_PROMPT")
	if card_prompt == "" {
		card_prompt = ImagePrompt
	}

	cb, err := webhook.ParseRequest(ChannelSecret, r)
	if err != nil {
		if err == linebot.ErrInvalidSignature {
			w.WriteHeader(400)
		} else {
			w.WriteHeader(500)
		}
		return
	}

	for _, event := range cb.Events {
		log.Printf("Got event %v", event)
		switch e := event.(type) {
		case webhook.MessageEvent:
			switch message := e.Message.(type) {
			// Handle only on text message
			case webhook.TextMessageContent:
				// 取得用戶 ID
				var uID string
				switch source := e.Source.(type) {
				case webhook.UserSource:
					uID = source.UserId
				case webhook.GroupSource:
					uID = source.UserId
				case webhook.RoomSource:
					uID = source.UserId
				}
				log.Println("Got text msg ID:", message.Id, " UID:", uID)
				userPath := fmt.Sprintf("%s/%s", DBCardPath, uID)
				fireDB.Path = userPath
				var People map[string]Person
				err := fireDB.GetFromDB(&People)
				if err != nil {
					log.Println("Error getting data from DB:", err)
				}

				// Marshall data to JSON
				jsonData, err := json.Marshal(People)
				if err != nil {
					fmt.Println("Error marshalling data to JSON:", err)
				}

				if message.Text == "list" {
					log.Println("Got list command")
					err = SendFlexMsg(e.ReplyToken, People, "名片列表")
					if err != nil {
						log.Println("Error send result", err)
					}
					continue
				} else {
					// Add Search prompt
					SearchPrompt := fmt.Sprintf("這是所有的名片資料，請根據輸入文字來查詢相關的名片資料 (%s)，例如: 名字, 職稱, 公司名稱。 查詢問句為： %s, 只要回覆我找到的 JSON Data", jsonData, message.Text)
					response := GeminiChatComplete(SearchPrompt, message.Text)
					log.Println("Find reply data:", response)

					// Remove first and last line,	which are the backticks.
					jsonData := removeFirstAndLastLine(response)

					var retPeople map[string]Person
					// unmarshall json to People
					err = json.Unmarshal([]byte(jsonData), &retPeople)
					if err != nil {
						fmt.Println("Unmarshal failed, ", err, "jsonData:", jsonData)
					}
					err = SendFlexMsg(e.ReplyToken, retPeople, "搜尋結果")
					if err != nil {
						log.Println("Error send result", err)
					}
				}

			// Handle only on Sticker message
			case webhook.StickerMessageContent:
				// log sticker id and package id.
				log.Printf("Got sticker message, packageID: %s, stickerID: %s", message.PackageId, message.StickerId)

			// Handle only image message
			case webhook.ImageMessageContent:
				// 取得用戶 ID
				var uID string
				switch source := e.Source.(type) {
				case webhook.UserSource:
					uID = source.UserId
				case webhook.GroupSource:
					uID = source.UserId
				case webhook.RoomSource:
					uID = source.UserId
				}
				userPath := fmt.Sprintf("%s/%s", DBCardPath, uID)
				fireDB.Path = userPath
				log.Println("Got img msg ID:", message.Id)
				//Get image binary from LINE server based on message ID.
				data, err := GetImageBinary(blob, message.Id)
				if err != nil {
					log.Println("Got GetMessageContent err:", err)
					continue
				}

				// Chat with Image
				ret, err := GeminiImage(data, card_prompt)
				if err != nil {
					ret = "無法辨識圖片內容文字，請重新輸入:" + err.Error()
					if err := replyText(e.ReplyToken, ret); err != nil {
						log.Print(err)
					}
					continue
				}

				log.Println("Got GeminiImage ret:", ret)

				// Remove first and last line,	which are the backticks.
				jsonData := removeFirstAndLastLine(ret)
				log.Println("Got jsonData:", jsonData)

				// Parse json and insert NotionDB
				var person Person
				err = json.Unmarshal([]byte(jsonData), &person)
				if err != nil {
					log.Println("Error parsing JSON:", err)
				}

				//TODO: Check email first before adding to firebase.
				if fireDB.SearchIfExist(person.Email) {
					log.Println("Email already exist in DB:", person.Email, "jsonData:", jsonData)
					if err := replyText(e.ReplyToken, "Email已經存在於資料庫:\n"+jsonData); err != nil {
						log.Print(err)
					}
					continue
				}

				// Add to firebase
				err = fireDB.InsertDB(person)
				if err != nil {
					log.Println("Error adding data to DB:", err)
				}
				People := map[string]Person{person.Name: person}
				if err := SendFlexMsg(e.ReplyToken, People, "新增到資料庫"); err != nil {
					log.Println("Error send result", err)
				}

			// Handle only video message
			case webhook.VideoMessageContent:
				log.Println("Got video msg ID:", message.Id)

			default:
				log.Printf("Unknown message: %v", message)
			}
		case webhook.PostbackEvent:
			log.Printf("Got postback: %v", e.Postback.Data)
		case webhook.JoinEvent:
			log.Printf("Got join event")
		case webhook.FollowEvent:
			log.Printf("message: Got followed event")
		case webhook.BeaconEvent:
			log.Printf("Got beacon: " + e.Beacon.Hwid)
		}
	}
}

// GetImageBinary: Get image binary from LINE server based on message ID.
func GetImageBinary(blob *messaging_api.MessagingApiBlobAPI, messageID string) ([]byte, error) {
	// Get image binary from LINE server based on message ID.
	content, err := blob.GetMessageContent(messageID)
	if err != nil {
		log.Println("Got GetMessageContent err:", err)
	}
	defer content.Body.Close()
	data, err := io.ReadAll(content.Body)
	if err != nil {
		log.Fatal(err)
	}

	return data, nil
}

// removeFirstAndLastLine takes a string and removes the first and last lines.
func removeFirstAndLastLine(s string) string {
	// Split the string into lines.
	lines := strings.Split(s, "\n")

	// If there are less than 3 lines, return an empty string because removing the first and last would leave nothing.
	if len(lines) < 3 {
		return ""
	}

	// Join the lines back together, skipping the first and last lines.
	return strings.Join(lines[1:len(lines)-1], "\n")
}
