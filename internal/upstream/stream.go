package upstream

import (
	"bufio"
	"encoding/json"
	"io"
	"strings"
)

type StreamEvent struct {
	ID           string
	ImageURL     string
	MessageID    string
	Result       string
	FinishReason string
}

type streamEnvelope struct {
	ID      string `json:"id"`
	Choices []struct {
		Delta struct {
			Content   string `json:"content"`
			MessageID string `json:"message_id"`
			Result    string `json:"result"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

// ParseStreamMessages reads Rita SSE payloads into a flat event list.
func ParseStreamMessages(r io.Reader) ([]StreamEvent, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	events := make([]StreamEvent, 0, 8)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "data:") {
			continue
		}

		raw := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if raw == "" || raw == "[DONE]" {
			continue
		}

		var envelope streamEnvelope
		if err := json.Unmarshal([]byte(raw), &envelope); err != nil {
			return nil, err
		}

		event := StreamEvent{ID: envelope.ID}
		if len(envelope.Choices) > 0 {
			choice := envelope.Choices[0]
			event.ImageURL = choice.Delta.Content
			event.MessageID = choice.Delta.MessageID
			event.Result = choice.Delta.Result
			event.FinishReason = choice.FinishReason
		}

		events = append(events, event)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return events, nil
}
