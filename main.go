// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	firebase "firebase.google.com/go"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"google.golang.org/api/option"
)

var bucketName string
var geminiKey string
var channelToken string
var ChannelSecret string

var bot *messaging_api.MessagingApiAPI
var blob *messaging_api.MessagingApiBlobAPI

func main() {
	ctx := context.Background()
	var err error
	geminiKey = os.Getenv("GOOGLE_GEMINI_API_KEY")
	channelToken = os.Getenv("ChannelAccessToken")
	ChannelSecret = os.Getenv("ChannelSecret")

	opt := option.WithCredentialsJSON([]byte(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")))
	config := &firebase.Config{DatabaseURL: os.Getenv("FIREBASE_URL")}
	app, err := firebase.NewApp(ctx, config, opt)
	if err != nil {
		log.Fatalf("error initializing firebase app: %v", err)
	}
	client, err := app.Database(ctx)
	if err != nil {
		log.Fatalf("error initializing database: %v", err)
	}
	fireDB.Client = client
	fireDB.CTX = ctx

	bot, err = messaging_api.NewMessagingApiAPI(channelToken)
	if err != nil {
		log.Fatal(err)
	}

	blob, err = messaging_api.NewMessagingApiBlobAPI(channelToken)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/callback", callbackHandler)
	port := os.Getenv("PORT")
	addr := fmt.Sprintf(":%s", port)
	http.ListenAndServe(addr, nil)
}
