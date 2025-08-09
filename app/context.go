package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"

	openai "github.com/sashabaranov/go-openai"
	"github.com/joho/godotenv"
)

var openaiClient *openai.Client

func main() {

    // Путь для сохранения файлов
	basePath := "/home/leouix/apps/test"

	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Ошибка загрузки .env файла")
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY не задан")
	}

	openaiClient = openai.NewClient(apiKey)

    currentDir, err := os.Getwd()
    if err != nil {
        log.Fatalf("Ошибка получения текущей директории: %v", err)
    }
    projectFiles := collectProjectFiles(currentDir, []string{".go"}, 5)

    // Чтение промпта из файла
	promptData, err := os.ReadFile("context.txt")
	if err != nil {
		log.Fatalf("Ошибка при чтении файла context.txt: %v", err)
	}
	prompt := string(promptData)

	// Отправка запроса в OpenAI
	response, err := openaiClient.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
        Model: openai.GPT3Dot5Turbo,
        ResponseFormat: &openai.ChatCompletionResponseFormat{
            Type: "json_object",
        },
        Messages: []openai.ChatCompletionMessage{
            {
                Role: "system",
                 Content: `Ты помощник, создающий структуру проекта.
                              Вот текущие файлы проекта:` + projectFiles + `
                              Твоя задача: вернуть **только** JSON вида:
                              {"./main.go": "...код...", "./README.md": "...текст..."}
                              Без комментариев, без Markdown, без текста до или после.
                              Если хочешь что-то пояснить — НЕ пиши этого. Верни только JSON.`,
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

	// content := response.Choices[0].Message.Content

    // Парсим JSON-ответ
    var files map[string]string

    raw := response.Choices[0].Message.Content
    re := regexp.MustCompile(`(?s)\{.*\}`)
    content := re.FindString(raw)
    if content == "" {
        log.Fatalf("Не удалось извлечь JSON из ответа: %s", raw)
    }
    err = json.Unmarshal([]byte(content), &files)




	err = json.Unmarshal([]byte(content), &files)
	if err != nil {
		log.Fatalf("Ошибка разбора JSON от OpenAI: %v\nRaw content: %s", err, content)
	}

    // Сохраняем файлы в указанную директорию
    for relPath, code := range files {
    	fullPath := filepath.Join(basePath, relPath)

    	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
    		log.Fatalf("Ошибка создания директорий для %s: %v", fullPath, err)
    	}

    	if err := os.WriteFile(fullPath, []byte(code), 0644); err != nil {
    		log.Fatalf("Ошибка записи файла %s: %v", fullPath, err)
    	}

    	fmt.Printf("Создан файл: %s\n", fullPath)
    }

    fmt.Println("✅ Все файлы успешно созданы.")

}


func collectProjectFiles(baseDir string, extensions []string, maxFiles int) string {
    var collected string
    count := 0

    filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return nil
        }
        // Пропуск директорий .git, node_modules и т.д.
        if info.IsDir() && (info.Name() == ".git" || info.Name() == "node_modules") {
            return filepath.SkipDir
        }
        if info.IsDir() {
            return nil
        }

        for _, ext := range extensions {
            if filepath.Ext(path) == ext && count < maxFiles {
                data, err := os.ReadFile(path)
                if err == nil {
                    collected += fmt.Sprintf("\n# %s\n%s\n", path, string(data))
                    count++
                }
                break
            }
        }
        return nil
    })

    return collected
}