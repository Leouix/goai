package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

    "github.com/joho/godotenv"
	openai "github.com/sashabaranov/go-openai"
)

var openaiClient *openai.Client

func main() {

    err := godotenv.Load("../.env")
    if err != nil {
        log.Fatal("Ошибка загрузки .env файла")
    }

	// rootDir := "."
	basePath := "/home/leouix/apps/goai"

	todos := findTodos(basePath, "// @todo")

	// Формируем JSON для OpenAI
	jsonData, err := json.MarshalIndent(todos, "", "  ")
	if err != nil {
		log.Fatalf("Ошибка кодирования в JSON: %v", err)
	}

    fmt.Printf("jsonData: %s\n", jsonData)

	// Запрос к OpenAI
	files := generateFilesFromOpenAI(string(jsonData))

	// Сохраняем результат
	saveFiles(basePath, files)
}

func findTodos(root, marker string) map[string]string {
	results := make(map[string]string)

	filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if d.IsDir() && (d.Name() == ".git" || d.Name() == "node_modules" || d.Name() == "vendor") {
			return filepath.SkipDir
		}

		if !d.IsDir() {
		    base := filepath.Base(path)
            if base == "todos.go" {
                return nil // пропускаем этот файл
            }

			data, err := os.ReadFile(path)
			if err != nil {
				return nil
			}
			content := string(data)

			if strings.Contains(content, marker) {
				newPath := path
				counter := 1
				for {
					if _, exists := results[newPath]; !exists {
						break
					}
					counter++
					newPath = fmt.Sprintf("%s#%d", path, counter)
				}
				results[newPath] = content
			}
		}
		return nil
	})

	return results
}

func generateFilesFromOpenAI(jsonData string) map[string]string {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY не задан")
	}

	client := openai.NewClient(apiKey)

	response, err := client.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo,
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: "json_object",
		},
		Messages: []openai.ChatCompletionMessage{
			{
				Role: "system",
				 Content: `Ты помощник программиста, мы пишем на PHP, Laravel, Symphony, Javascript, Go, используешь Docker, SQL и прочие технологии. Тебе предоставляется запрос вида: {'путь к   файлу': 'страница'.
                  Твоя задача: выполнить @todo, указанный на странице и заменить @todo своим кодом. Например удалить строку '// @todo выполнить проверку переменной $var' и на этом месте написать if (isset($var)) ...

                  Вернуть **только** JSON вида:
                  {"путь к   файлу": "...исправленный код файла..."}
                  путь к файлу нужно оставить неизменным.
                  Без комментариев, без Markdown, без текста до или после.
                  Если хочешь что-то пояснить — НЕ пиши этого. Верни только JSON.`,
			},
			{
				Role:    "user",
				Content: string(jsonData),
			},
		},
	})

	if err != nil {
		log.Fatalf("Ошибка OpenAI запроса: %v", err)
	}

	raw := response.Choices[0].Message.Content

	// Страховка на случай, если модель всё же добавит текст
	re := regexp.MustCompile(`(?s)\{.*\}`)
	content := re.FindString(raw)
	if content == "" {
		log.Fatalf("Не удалось извлечь JSON из ответа: %s", raw)
	}

	var files map[string]string
	err = json.Unmarshal([]byte(content), &files)
	if err != nil {
		log.Fatalf("Ошибка разбора JSON: %v\nRaw content: %s", err, content)
	}

	return files
}

func saveFiles(basePath string, files map[string]string) {
	for relPath, code := range files {
		fullPath := filepath.Join(relPath)

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
