package main

import (
	"context"
	"fmt"
	"log"
"errors"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

func GeminiImage(imgData []byte, prompt string) (string, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(geminiKey))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-pro-vision")
	model.SafetySettings = []*genai.SafetySetting{
		{
			Category:  genai.HarmCategoryUnspecified,
			Threshold: genai.HarmBlockNone,
		},
		{
			Category:  genai.HarmCategoryHarassment,
			Threshold: genai.HarmBlockNone,
		},
		{
			Category:  genai.HarmCategoryViolence,
			Threshold: genai.HarmBlockNone,
		},
		{
			Category:  genai.HarmCategorySexualContent,
			Threshold: genai.HarmBlockNone,
		},
	}
	value := float32(0.5)
	model.Temperature = &value
	data := []genai.Part{
		genai.ImageData("png", imgData),
		genai.Text(prompt),
	}
	log.Println("Begin processing image...")
	resp, err := model.GenerateContent(ctx, data...)
	log.Println("Finished processing image...", resp)
	if err != nil {
		log.Println(err)
		var berr *genai.BlockedError
		if errors.As(err, &berr) {
			log.Println("Blocked error:", resp.PromptFeedback)
		}
		return "", err
	}

	return printResponse(resp), nil
}

// Gemini Chat Complete: Iput a prompt and get the response string.
func GeminiChatComplete(prompt, req string) string {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(geminiKey))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()
	model := client.GenerativeModel("gemini-1.5-flash-latest")
	value := float32(0.8)
	model.Temperature = &value
	res, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		log.Fatal(err)
	}
	return printResponse(res)
}

func printResponse(resp *genai.GenerateContentResponse) string {
	var ret string
	for _, cand := range resp.Candidates {
		for _, part := range cand.Content.Parts {
			ret = ret + fmt.Sprintf("%v", part)
			fmt.Println(part)
		}
	}
	return ret
}
