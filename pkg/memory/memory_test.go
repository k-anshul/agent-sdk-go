package memory

import (
	"context"
	"testing"

	"github.com/pontus-devoteam/agent-sdk-go/pkg/result"
)

func TestInMemoryStorage_Basic(t *testing.T) {
	ctx := context.Background()
	memory := NewInMemoryStorage()
	sessionID := "test-session"

	// Test initial state
	size, err := memory.Size(ctx, sessionID)
	if err != nil {
		t.Fatalf("Failed to get size: %v", err)
	}
	if size != 0 {
		t.Errorf("Expected size 0, got %d", size)
	}

	// Create test run result
	runResult := &result.RunResult{
		Input: "test input",
		NewItems: []result.RunItem{
			&result.MessageItem{
				Role:    "user",
				Content: "Hello",
			},
			&result.MessageItem{
				Role:    "assistant",
				Content: "Hi there!",
			},
		},
		FinalOutput: "Hi there!",
	}

	// Add to memory
	err = memory.Add(ctx, sessionID, runResult)
	if err != nil {
		t.Fatalf("Failed to add to memory: %v", err)
	}

	// Check size
	size, err = memory.Size(ctx, sessionID)
	if err != nil {
		t.Fatalf("Failed to get size: %v", err)
	}
	if size != 3 { // 2 from NewItems + 1 from FinalOutput
		t.Errorf("Expected size 3, got %d", size)
	}

	// Get all items
	items, err := memory.GetAll(ctx, sessionID)
	if err != nil {
		t.Fatalf("Failed to get all items: %v", err)
	}
	if len(items) != 3 {
		t.Errorf("Expected 3 items, got %d", len(items))
	}

	// Check item types
	if items[0].GetType() != "message" {
		t.Errorf("Expected first item type 'message', got '%s'", items[0].GetType())
	}
	if items[1].GetType() != "message" {
		t.Errorf("Expected second item type 'message', got '%s'", items[1].GetType())
	}
	if items[2].GetType() != "message" {
		t.Errorf("Expected third item type 'message', got '%s'", items[2].GetType())
	}
}

func TestInMemoryStorage_GetWithCriteria(t *testing.T) {
	ctx := context.Background()
	memory := NewInMemoryStorage()
	sessionID := "test-session"

	// Add various types of items
	runResult := &result.RunResult{
		NewItems: []result.RunItem{
			&result.MessageItem{Role: "user", Content: "Message 1"},
			&result.ToolCallItem{Name: "tool1", Parameters: map[string]interface{}{"param": "value"}},
			&result.MessageItem{Role: "assistant", Content: "Message 2"},
			&result.ToolResultItem{Name: "tool1", Result: "result"},
			&result.HandoffItem{AgentName: "Agent1", Input: "handoff input"},
		},
	}

	err := memory.Add(ctx, sessionID, runResult)
	if err != nil {
		t.Fatalf("Failed to add to memory: %v", err)
	}

	// Test filtering by type
	messageItems, err := memory.GetByType(ctx, sessionID, []string{"message"})
	if err != nil {
		t.Fatalf("Failed to get message items: %v", err)
	}
	if len(messageItems) != 2 {
		t.Errorf("Expected 2 message items, got %d", len(messageItems))
	}

	// Test limit
	limitedItems, err := memory.Get(ctx, sessionID, &GetCriteria{Limit: 3})
	if err != nil {
		t.Fatalf("Failed to get limited items: %v", err)
	}
	if len(limitedItems) != 3 {
		t.Errorf("Expected 3 limited items, got %d", len(limitedItems))
	}

	// Test reverse order
	recentItems, err := memory.GetRecent(ctx, sessionID, 2)
	if err != nil {
		t.Fatalf("Failed to get recent items: %v", err)
	}
	if len(recentItems) != 2 {
		t.Errorf("Expected 2 recent items, got %d", len(recentItems))
	}
	// Should be in reverse order (newest first)
	if recentItems[0].GetType() != "handoff" {
		t.Errorf("Expected first recent item to be 'handoff', got '%s'", recentItems[0].GetType())
	}
}

func TestInMemoryStorage_Clear(t *testing.T) {
	ctx := context.Background()
	memory := NewInMemoryStorage()
	sessionID := "test-session"

	// Add some items
	runResult := &result.RunResult{
		NewItems: []result.RunItem{
			&result.MessageItem{Role: "user", Content: "Test"},
		},
	}

	err := memory.Add(ctx, sessionID, runResult)
	if err != nil {
		t.Fatalf("Failed to add to memory: %v", err)
	}

	// Verify items exist
	size, _ := memory.Size(ctx, sessionID)
	if size != 1 {
		t.Errorf("Expected size 1 before clear, got %d", size)
	}

	// Clear memory
	err = memory.Clear(ctx, sessionID)
	if err != nil {
		t.Fatalf("Failed to clear memory: %v", err)
	}

	// Verify memory is empty
	size, _ = memory.Size(ctx, sessionID)
	if size != 0 {
		t.Errorf("Expected size 0 after clear, got %d", size)
	}
}

func TestInMemoryStorage_ConcurrentAccess(t *testing.T) {
	ctx := context.Background()
	memory := NewInMemoryStorage()
	sessionID := "test-session"

	// Test concurrent reads and writes
	done := make(chan bool, 2)

	// Writer goroutine
	go func() {
		for i := 0; i < 10; i++ {
			runResult := &result.RunResult{
				NewItems: []result.RunItem{
					&result.MessageItem{Role: "user", Content: "Concurrent message"},
				},
			}
			memory.Add(ctx, sessionID, runResult)
		}
		done <- true
	}()

	// Reader goroutine
	go func() {
		for i := 0; i < 10; i++ {
			memory.GetAll(ctx, sessionID)
			memory.Size(ctx, sessionID)
		}
		done <- true
	}()

	// Wait for both goroutines to complete
	<-done
	<-done

	// Verify final state
	size, err := memory.Size(ctx, sessionID)
	if err != nil {
		t.Fatalf("Failed to get final size: %v", err)
	}
	if size != 10 {
		t.Errorf("Expected final size 10, got %d", size)
	}
}

func TestInMemoryStorage_SessionIsolation(t *testing.T) {
	ctx := context.Background()
	memory := NewInMemoryStorage()
	sessionID1 := "session-1"
	sessionID2 := "session-2"

	// Add items to session 1
	runResult1 := &result.RunResult{
		NewItems: []result.RunItem{
			&result.MessageItem{Role: "user", Content: "Session 1 message"},
		},
	}
	err := memory.Add(ctx, sessionID1, runResult1)
	if err != nil {
		t.Fatalf("Failed to add to session 1: %v", err)
	}

	// Add items to session 2
	runResult2 := &result.RunResult{
		NewItems: []result.RunItem{
			&result.MessageItem{Role: "user", Content: "Session 2 message"},
			&result.MessageItem{Role: "assistant", Content: "Session 2 response"},
		},
	}
	err = memory.Add(ctx, sessionID2, runResult2)
	if err != nil {
		t.Fatalf("Failed to add to session 2: %v", err)
	}

	// Check session 1 size
	size1, err := memory.Size(ctx, sessionID1)
	if err != nil {
		t.Fatalf("Failed to get session 1 size: %v", err)
	}
	if size1 != 1 {
		t.Errorf("Expected session 1 size 1, got %d", size1)
	}

	// Check session 2 size
	size2, err := memory.Size(ctx, sessionID2)
	if err != nil {
		t.Fatalf("Failed to get session 2 size: %v", err)
	}
	if size2 != 2 {
		t.Errorf("Expected session 2 size 2, got %d", size2)
	}

	// Clear session 1
	err = memory.Clear(ctx, sessionID1)
	if err != nil {
		t.Fatalf("Failed to clear session 1: %v", err)
	}

	// Verify session 1 is empty but session 2 is unchanged
	size1, _ = memory.Size(ctx, sessionID1)
	if size1 != 0 {
		t.Errorf("Expected session 1 size 0 after clear, got %d", size1)
	}

	size2, _ = memory.Size(ctx, sessionID2)
	if size2 != 2 {
		t.Errorf("Expected session 2 size still 2 after clearing session 1, got %d", size2)
	}

	// Test GetSessions
	sessions, err := memory.GetSessions(ctx)
	if err != nil {
		t.Fatalf("Failed to get sessions: %v", err)
	}
	if len(sessions) != 1 {
		t.Errorf("Expected 1 active session, got %d", len(sessions))
	}
	if sessions[0] != sessionID2 {
		t.Errorf("Expected active session to be %s, got %s", sessionID2, sessions[0])
	}
}
