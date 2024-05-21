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

	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
)

var geminiKey string
var ChannelSecret string

var bot *messaging_api.MessagingApiAPI
var blob *messaging_api.MessagingApiBlobAPI

func main() {
	ctx := context.Background()
	var err error

	// Get the environment variables
	geminiKey = os.Getenv("GOOGLE_GEMINI_API_KEY")
	channelToken := os.Getenv("ChannelAccessToken")
	ChannelSecret = os.Getenv("ChannelSecret")
	gap := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	firebaseURL := os.Getenv("FIREBASE_URL")

	// Initialize Firebase
	initFirebase(gap, firebaseURL, ctx)

	// Initialize LINE Bot
	bot, err = messaging_api.NewMessagingApiAPI(channelToken)
	if err != nil {
		log.Fatal(err)
	}

	blob, err = messaging_api.NewMessagingApiBlobAPI(channelToken)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", callbackHandler)
	port := os.Getenv("PORT")
	addr := fmt.Sprintf(":%s", port)
	http.ListenAndServe(addr, nil)
}
