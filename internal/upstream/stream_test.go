package upstream

import (
	"strings"
	"testing"
)

func TestParseStreamMessagesReadsResultAndStop(t *testing.T) {
	t.Parallel()

	payload := strings.NewReader(`data: {"choices":[{"delta":{"content":"","message_id":"","result":""},"finish_reason":"","index":0}],"id":"AiArt-parent-1"}

data: {"choices":[{"delta":{"content":"https://img.example/result.png","message_id":"parent-1-0","result":"success"},"finish_reason":"success","index":0}],"id":"AiArt-parent-1"}

data: {"choices":[{"delta":{},"finish_reason":"stop","index":0}],"id":"AiArt-parent-1"}

`)

	events, err := ParseStreamMessages(payload)
	if err != nil {
		t.Fatalf("ParseStreamMessages() error = %v", err)
	}

	if len(events) != 3 {
		t.Fatalf("event count = %d, want 3", len(events))
	}

	if events[1].ImageURL != "https://img.example/result.png" {
		t.Fatalf("events[1].ImageURL = %q, want result url", events[1].ImageURL)
	}

	if events[1].Result != "success" {
		t.Fatalf("events[1].Result = %q, want success", events[1].Result)
	}

	if events[2].FinishReason != "stop" {
		t.Fatalf("events[2].FinishReason = %q, want stop", events[2].FinishReason)
	}
}

func TestParseStreamMessagesRejectsBrokenJSON(t *testing.T) {
	t.Parallel()

	payload := strings.NewReader("data: {broken}\n\n")

	if _, err := ParseStreamMessages(payload); err == nil {
		t.Fatal("ParseStreamMessages() error = nil, want json parse error")
	}
}
