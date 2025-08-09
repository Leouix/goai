package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

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
	promptData, err := os.ReadFile("start.txt")
	if err != nil {
		log.Fatalf("Ошибка при чтении файла start.txt: %v", err)
	}
	prompt := string(promptData)

	// @todos записывать в лог строку prompt

	// Отправка запроса в OpenAI
	response, err := openaiClient.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{
				Role: "system",
				Content: `Ты помощник, создающий структуру файлов. Ответ должен быть в формате JSON: путь => содержимое. Сгенерируй JSON-объект, где ключи — это пути к файлам, а значения — содержимое. Например:
{
  "./main.go": "package main\n\nfunc main() {...}",
  "./README.md": "# My Project"
}`,            },
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

	// Парсим JSON-ответ
	var files map[string]string
	err = json.Unmarshal([]byte(content), &files)
	if err != nil {
		log.Fatalf("Ошибка разбора JSON от OpenAI: %v\nRaw content: %s", err, content)
	}

	// Путь для сохранения файлов
	basePath := "/home/leouix/apps/test"

	// Добавить проверку что basePath существует
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		log.Fatalf("Директория %s не существует", basePath)
	}

	for relPath, code := range files {
		fullPath := filepath.Join(basePath, relPath)

		// @todos записывать в лог fullPath

		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			log.Fatalf("Ошибка создания директорий: %v", err)
		}

		if err := os.WriteFile(fullPath, []byte(code), 0644); err != nil {
			log.Fatalf("Ошибка записи файла: %v", err)
		}

		fmt.Printf("Создан файл: %s\n", fullPath)
	}
}
