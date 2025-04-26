package backend

import (
	"fmt"
	"os"
	"strings"
)

type Setting struct{}

func NewSetting() *Setting {
	return &Setting{}
}

func (s *Setting) SaveOpenAIKey(key string) error {
	// Read the current .env file
	content, err := os.ReadFile(".env")
	if err != nil {
		if os.IsNotExist(err) {
			// Create new .env file with the API key
			var b []byte
			b = fmt.Appendf(b, "OPENAI_API_KEY=%s\n", key)
			err = os.WriteFile(".env", b, 0644)
			if err != nil {
				return fmt.Errorf("failed to create .env file: %w", err)
			}
			return nil
		}
		return fmt.Errorf("failed to read .env file: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	keyFound := false
	newLines := make([]string, 0)

	// Look for existing OPENAI_API_KEY
	for _, line := range lines {
		if strings.HasPrefix(line, "OPENAI_API_KEY=") {
			newLines = append(newLines, fmt.Sprintf("OPENAI_API_KEY=%s", key))
			keyFound = true
		} else if line != "" {
			newLines = append(newLines, line)
		}
	}

	// If key wasn't found, add it
	if !keyFound {
		newLines = append(newLines, fmt.Sprintf("OPENAI_API_KEY=%s", key))
	}

	// Write back to .env file
	err = os.WriteFile(".env", []byte(strings.Join(newLines, "\n")), 0644)
	if err != nil {
		return fmt.Errorf("failed to write .env file: %w", err)
	}

	return nil
}

func (s *Setting) GetOpenAIKey() (string, error) {
	content, err := os.ReadFile(".env")
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("failed to read .env file: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "OPENAI_API_KEY=") {
			return strings.TrimPrefix(line, "OPENAI_API_KEY="), nil
		}
	}

	return "", nil
}
