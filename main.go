package main

import (
    "context"
    "net/http"
    "encoding/json"
    "os"
    "path/filepath"

    "github.com/gin-gonic/gin"
    openai "github.com/sashabaranov/go-openai"
)

type RequestBody struct {
    Prompt string `json:"prompt"`
    BasePath string `json:"basePath"`
}

var openaiClient *openai.Client

func main() {

    r := gin.Default()
    r.POST("/generate", handleGenerate)
    r.Run(":8080")
}

func handleGenerate(c *gin.Context) {
    var req RequestBody
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
        return
    }

    response, err := openaiClient.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
       Model: openai.GPT3Dot5Turbo,
       Messages: []openai.ChatCompletionMessage{
           {
               Role: "system",
               Content: `Ты помощник, создающий структуру файлов. Ответ должен быть в формате JSON: путь => содержимое. Сгенерируй JSON-объект, где ключи — это пути к файлам, а значения
            — содержимое. Например:
       {
         "./main.go": "package main\n\nfunc main() {...}",
         "./README.md": "# My Project"
       }`,
               },
               {
                   Role:    "user",
                   Content: req.Prompt,
               },
           },
       })


    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    content := response.Choices[0].Message.Content

    // Допустим, OpenAI вернул JSON-объект: map[string]string
    var files map[string]string
    err = json.Unmarshal([]byte(content), &files)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid JSON from OpenAI", "raw": content})
        return
    }

    // Задаём базовый путь по умолчанию
	basePath := "/goai-tmp"

   // Сохраняем файлы в указанную директорию
   	for relPath, code := range files {
   		fullPath := filepath.Join(basePath, relPath)

   		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
   			c.JSON(http.StatusInternalServerError, gin.H{
   				"error": "Could not create directories",
   				"path":  fullPath,
   			})
   			return
   		}

   		if err := os.WriteFile(fullPath, []byte(code), 0644); err != nil {
   			c.JSON(http.StatusInternalServerError, gin.H{
   				"error": "Could not write file",
   				"path":  fullPath,
   			})
   			return
   		}
   	}

    c.JSON(http.StatusOK, gin.H{
    		"status": "success",
    		"files":  files})
}
