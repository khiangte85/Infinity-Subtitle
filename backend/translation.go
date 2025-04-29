package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"infinity-subtitle/backend/logger"

	_ "github.com/joho/godotenv/autoload"
	"github.com/sashabaranov/go-openai"
	"golang.org/x/exp/maps"
	"golang.org/x/time/rate"
	"runtime"
)

type TranslationService struct {
	client      *openai.Client
	rateLimiter *rate.Limiter
	logger      *logger.Logger
}

func NewTranslationService() (*TranslationService, error) {
	log, err := logger.GetLogger()
	if err != nil {
		return nil, fmt.Errorf("failed to get logger: %w", err)
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	log.Info("Initializing translation service with API key: %s", apiKey)

	return &TranslationService{
		client:      openai.NewClient(apiKey),
		rateLimiter: rate.NewLimiter(rate.Every(time.Second/60), 1),
		logger:      log,
	}, nil
}

// Close should be called when the service is no longer needed
func (ts *TranslationService) Close() error {
	return nil
}

type TranslationRequest struct {
	Texts          map[string]string `json:"texts"`
	SourceLanguage string            `json:"source_language"`
	TargetLanguage string            `json:"target_language"`
}

type TranslationResponse struct {
	Translations map[string]string `json:"translations"`
}

func (ts *TranslationService) translate(ctx context.Context, texts map[string]string, sourceLang, targetLang string) (map[string]string, error) {
	if len(texts) == 0 {
		return make(map[string]string), nil
	}

	if err := ts.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	prompt := fmt.Sprintf("Translate the following texts from %s to %s.\nTexts to translate:\n", sourceLang, targetLang)
	for text := range texts {
		prompt += fmt.Sprintf("- %s\n", text)
	}

	prompt += "\nOutput format should be a JSON string that can be Unmarshalable by Golang with the same keys as input, " +
		"containing only the translations. Do not include ```json ```."

	ts.logger.Info("Translation prompt: %s", prompt)

	resp, err := ts.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT4oMini,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)
	if err != nil {
		ts.logger.Error("Chat completion error: %v", err)
		return nil, fmt.Errorf("failed to create chat completion: %w", err)
	}

	if len(resp.Choices) == 0 {
		ts.logger.Error("No response from OpenAI: %v", resp.Choices)
		return nil, fmt.Errorf("no response from OpenAI")
	}

	var translations map[string]string
	err = json.Unmarshal([]byte(resp.Choices[0].Message.Content), &translations)
	if err != nil {
		ts.logger.Error("Failed to parse translation response: %s", resp.Choices[0].Message.Content)
		return nil, fmt.Errorf("failed to parse translation response: %w", err)
	}

	return translations, nil
}

func (ts *TranslationService) processBatch(ctx context.Context, texts map[string]string, sourceLang, targetLang string) map[string]string {
	var wg sync.WaitGroup
	results := make(map[string]string)
	mu := sync.Mutex{}

	// Split texts into batches of 20
	batchSize := 20
	batches := make([]map[string]string, 0)
	currentBatch := make(map[string]string)

	for text := range texts {
		currentBatch[text] = ""
		if len(currentBatch) >= batchSize {
			batches = append(batches, currentBatch)
			currentBatch = make(map[string]string)
		}
	}
	if len(currentBatch) > 0 {
		batches = append(batches, currentBatch)
	}

	// Fan-out: Create multiple workers
	workerCount := runtime.NumCPU()
	batchChan := make(chan map[string]string, len(batches))
	resultChan := make(chan map[string]string, len(batches))

	// Start workers
	for range workerCount {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for batch := range batchChan {
				translations, err := ts.translate(ctx, batch, sourceLang, targetLang)
				if err == nil {
					resultChan <- translations
				} 
			}
		}()
	}

	// Send batches to workers
	for _, batch := range batches {
		batchChan <- batch
	}
	close(batchChan)

	// Fan-in: Collect results
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Process results
	for translations := range resultChan {
		mu.Lock()
		maps.Copy(results, translations)
		mu.Unlock()
	}

	ts.logger.Info("Completed translations: %+v", results)

	return results
}
