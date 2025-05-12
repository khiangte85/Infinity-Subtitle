package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"infinity-subtitle/backend/logger"

	"runtime"

	_ "github.com/joho/godotenv/autoload"
	"github.com/sashabaranov/go-openai"
	"golang.org/x/time/rate"
)

type TextToTranslate struct {
	ID          int    `json:"id"`
	SourceText  string `json:"source_text"`
	Translation string `json:"translation"`
}

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

func (ts *TranslationService) translate(ctx context.Context, textsToTranslate []TextToTranslate,
	sourceLang, targetLang string) ([]TextToTranslate, error) {
	if len(textsToTranslate) == 0 {
		return make([]TextToTranslate, 0), nil
	}

	if err := ts.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	prompt := fmt.Sprintf("Translate the following data from %s to %s: use `SourceText` value as input "+
		" and put output value to `Translation`.\n", sourceLang, targetLang)
	for _, text := range textsToTranslate {
		prompt += fmt.Sprintf("%+v\n", text)
	}

	prompt += "\nOutput format should be ARRAY of JSON. Output format should be \n" +
		"[{\"id\": 1, \"source_text\": \"\", \"translation\": \"\"}, {\"id\": 2, \"source_text\": \"\", \"translation\": \"\"}] " +
		"\nDO NOT INCLUDE ```json"

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

	ts.logger.Info("Translation response: %s", resp.Choices[0].Message.Content)

	var translations []TextToTranslate
	err = json.Unmarshal([]byte(resp.Choices[0].Message.Content), &translations)
	if err != nil {
		ts.logger.Error("Failed to parse translation response: %s", resp.Choices[0].Message.Content)
		return nil, fmt.Errorf("failed to parse translation response: %w", err)
	}

	return translations, nil
}

func (ts *TranslationService) processBatch(ctx context.Context, textsToTranslate []TextToTranslate,
	sourceLang, targetLang string) []TextToTranslate {
	ts.logger.Info("Processing batch of %d texts", len(textsToTranslate))

	var wg sync.WaitGroup
	results := make([]TextToTranslate, 0)
	mu := sync.Mutex{}

	// Split texts into batches of 20
	batchSize := 20
	batches := make([][]TextToTranslate, 0)
	currentBatch := make([]TextToTranslate, 0)

	for _, textToTranslate := range textsToTranslate {
		currentBatch = append(currentBatch, textToTranslate)
		if len(currentBatch) >= batchSize {
			batches = append(batches, currentBatch)
			currentBatch = make([]TextToTranslate, 0)
		}
	}
	if len(currentBatch) > 0 {
		batches = append(batches, currentBatch)
	}

	// Fan-out: Create multiple workers
	workerCount := runtime.NumCPU()
	batchChan := make(chan []TextToTranslate, len(batches))
	resultChan := make(chan []TextToTranslate, len(batches))

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
		results = append(results, translations...)
		mu.Unlock()
	}

	ts.logger.Info("Completed translations: %+v", results)

	return results
}
