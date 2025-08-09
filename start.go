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

func main() {...}", "./README.md": "# My Project" }`, }, { Role: "user", Content: prompt, }, }, }) if err != nil { log.Fatalf("Ошибка OpenAI запроса: %v", err) } content := response.Choices[0].Message.Content var files map[string]string err = json.Unmarshal([]byte(content), &files) if err != nil { log.Fatalf("Ошибка разбора JSON от OpenAI: %v
Raw content: %s", err, content) } basePath := "/home/leouix/apps/test" if _, err := os.Stat(basePath); os.IsNotExist(err) { log.Fatalf("Директория %s не существует", basePath) } for relPath, code := range files { fullPath := filepath.Join(basePath, relPath) // Комментарий о записи в лог fullPath if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil { log.Fatalf("Ошибка создания директорий: %v", err) } if err := os.WriteFile(fullPath, []byte(code), 0644); err != nil { log.Fatalf("Ошибка записи файла: %v", err) } fmt.Printf("Создан файл: %s
", fullPath) } }