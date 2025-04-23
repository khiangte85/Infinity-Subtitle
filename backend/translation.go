package backend

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/sashabaranov/go-openai"
)

type TranslationService struct {
	client *openai.Client
}

func NewTranslationService() *TranslationService {
	return &TranslationService{
		client: openai.NewClient(os.Getenv("OPENAI_API_KEY")),
	}
}

type TranslationRequest struct {
	Text           string `json:"text"`
	SourceLanguage string `json:"source_language"`
	TargetLanguage string `json:"target_language"`
}

type TranslationResponse struct {
	Text string `json:"text"`
}

func (ts *TranslationService) translate(ctx context.Context, text, sourceLang, targetLang string) (string, error) {
	if text == "" {
		return "", nil
	}

	prompt := fmt.Sprintf("Translate the following text from %s to %s. Only output the translation, nothing else.\n\nText to translate: %s", sourceLang, targetLang, text)

	resp, err := ts.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to create chat completion: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}

	return resp.Choices[0].Message.Content, nil
}

func (ts *TranslationService) processBatch(ctx context.Context, texts []string, sourceLang, targetLang string) map[string]string {
	var wg sync.WaitGroup
	results := make(map[string]string)
	mu := sync.Mutex{}

	// Fan-out: Create multiple workers
	workerCount := 5
	textChan := make(chan string, len(texts))
	resultChan := make(chan struct {
		original   string
		translated string
	}, len(texts))

	// Start workers
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for text := range textChan {
				translated, err := ts.translate(ctx, text, sourceLang, targetLang)
				if err == nil {
					resultChan <- struct {
						original   string
						translated string
					}{text, translated}
				}
			}
		}()
	}

	// Send texts to workers
	for _, text := range texts {
		textChan <- text
	}
	close(textChan)

	// Fan-in: Collect results
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Process results
	for result := range resultChan {
		mu.Lock()
		results[result.original] = result.translated
		mu.Unlock()
	}

	return results
}
