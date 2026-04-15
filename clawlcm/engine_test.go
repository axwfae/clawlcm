package clawlcm

import (
	"context"
	"os"
	"testing"

	"github.com/axwfae/clawlcm/logger"
)

func TestEngineBootstrap(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "clawlcm-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	cfg := DefaultConfig()
	cfg.DatabasePath = tmpFile.Name()
	cfg.SummaryModel = ""

	engine, err := NewEngine(cfg, logger.New())
	if err != nil {
		t.Fatalf("NewEngine failed: %v", err)
	}

	resp, err := engine.Bootstrap(context.Background(), BootstrapRequest{
		SessionKey:  "test:session:1",
		SessionID:   "uuid-1",
		TokenBudget: 128000,
		Messages: []Message{
			{Role: "user", Content: "Hello"},
			{Role: "assistant", Content: "Hi there!"},
		},
	})
	if err != nil {
		t.Fatalf("Bootstrap failed: %v", err)
	}

	if resp.ConversationID == 0 {
		t.Error("Expected non-zero conversation ID")
	}

	if len(resp.Messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(resp.Messages))
	}
}

func TestEngineIngest(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "clawlcm-test-ingest-*.db")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	cfg := DefaultConfig()
	cfg.DatabasePath = tmpFile.Name()

	engine, err := NewEngine(cfg, logger.New())
	if err != nil {
		t.Fatal(err)
	}

	engine.Bootstrap(context.Background(), BootstrapRequest{
		SessionKey:  "test:session:2",
		SessionID:   "uuid-2",
		TokenBudget: 128000,
		Messages:    []Message{},
	})

	resp, err := engine.Ingest(context.Background(), IngestRequest{
		SessionKey:  "test:session:2",
		SessionID:   "uuid-2",
		TokenBudget: 128000,
		Message:     Message{Role: "user", Content: "New message"},
	})
	if err != nil {
		t.Fatalf("Ingest failed: %v", err)
	}

	if resp.MessageID == 0 {
		t.Error("Expected non-zero message ID")
	}

	if resp.Ordinal != 0 {
		t.Errorf("Expected ordinal 0, got %d", resp.Ordinal)
	}
}

func TestEngineAssemble(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "clawlcm-test-assemble-*.db")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	cfg := DefaultConfig()
	cfg.DatabasePath = tmpFile.Name()

	engine, err := NewEngine(cfg, logger.New())
	if err != nil {
		t.Fatal(err)
	}

	engine.Bootstrap(context.Background(), BootstrapRequest{
		SessionKey:  "test:session:3",
		SessionID:   "uuid-3",
		TokenBudget: 128000,
		Messages: []Message{
			{Role: "user", Content: "Message 1"},
			{Role: "assistant", Content: "Message 2"},
			{Role: "user", Content: "Message 3"},
			{Role: "assistant", Content: "Message 4"},
		},
	})

	resp, err := engine.Assemble(context.Background(), AssembleRequest{
		SessionKey:  "test:session:3",
		TokenBudget: 128000,
	})
	if err != nil {
		t.Fatalf("Assemble failed: %v", err)
	}

	if len(resp.Messages) == 0 {
		t.Error("Expected non-empty messages")
	}

	if resp.Stats.RawMessageCount != 4 {
		t.Errorf("Expected 4 raw messages, got %d", resp.Stats.RawMessageCount)
	}
}

func TestEngineInfo(t *testing.T) {
	cfg := DefaultConfig()
	cfg.DatabasePath = ":memory:"

	engine, err := NewEngine(cfg, logger.New())
	if err != nil {
		t.Fatal(err)
	}

	info := engine.Info()
	if info.ID != "clawlcm" {
		t.Errorf("Expected ID 'clawlcm', got '%s'", info.ID)
	}
	if info.Version != "0.3.0" {
		t.Errorf("Expected version '0.3.0', got '%s'", info.Version)
	}
}
