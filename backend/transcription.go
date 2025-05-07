package backend

import (
	"context"
	"fmt"
	"infinity-subtitle/backend/logger"
	"os"

	openai "github.com/sashabaranov/go-openai"
)

const (
	modelName = "whisper-1"
)

func transcribeAudio(audioPath string, language string) (string, error) {
	// Create OpenAI client
	openaiClient := openai.NewClient(os.Getenv("OPENAI_API_KEY"))

	logger, err := logger.GetLogger()
	if err != nil {
		return "", fmt.Errorf("failed to create logger: %w", err)
	}

	// Open the audio file
	audioFile, err := os.Open(audioPath)
	if err != nil {
		return "", fmt.Errorf("failed to open audio file: %w", err)
	}
	defer audioFile.Close()

	// Create transcription request
	req := openai.AudioRequest{
		Model:    modelName,
		FilePath: audioPath,
		Language: language,
		Prompt:   "Please transcribe the audio file to srt format",
		Format:   openai.AudioResponseFormatSRT,
	}

	// Send request to OpenAI
	logger.Info("Transcribing audio file: %s", audioPath)
	resp, err := openaiClient.CreateTranscription(context.Background(), req)
	if err != nil {
		return "", fmt.Errorf("failed to create transcription: %w", err)
	}

	logger.Info("Transcription completed: %+v", resp.Text)

	return resp.Text, nil
}
