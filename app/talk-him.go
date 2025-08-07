package main

import (
	"context"
	"fmt"
	"log"
	"os"

	openai "github.com/sashabaranov/go-openai"
	"github.com/joho/godotenv"
)

var openaiClient *openai.Client

func main() {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Ошибка загрузки .env файла")
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY не задан")
	}

	openaiClient = openai.NewClient(apiKey)

	// Чтение промпта из файла
	promptData, err := os.ReadFile("talk-him.txt")
	if err != nil {
		log.Fatalf("Ошибка при чтении файла talk-him.txt: %v", err)
	}
	prompt := string(promptData)

	// Отправка запроса в OpenAI
	response, err := openaiClient.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{
				Role: "system",
				Content: `Ты помощник-консультант, с которым мы создаем систему на Go lang, PHP, Docker, VueJs и прочее. Просто старайся следовать документации по вопросам`,
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
	})

	if err != nil {
		log.Fatalf("Ошибка OpenAI запроса: %v", err)
	}

	content := response.Choices[0].Message.Content

	fmt.Println("Ответ от Oscara:\n")
    fmt.Println(content)
}
