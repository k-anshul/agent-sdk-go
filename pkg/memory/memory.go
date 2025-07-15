package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/pontus-devoteam/agent-sdk-go/pkg/result"
)

// Memory interface defines the contract for memory implementations
type Memory interface {
	// Add adds a run result to memory for a specific session
	Add(ctx context.Context, sessionID string, runResult *result.RunResult) error

	// Get retrieves memory items based on criteria for a specific session
	Get(ctx context.Context, sessionID string, criteria *GetCriteria) ([]result.RunItem, error)

	// Clear clears all memory for a specific session
	Clear(ctx context.Context, sessionID string) error

	// ClearAll clears all memory for all sessions
	ClearAll(ctx context.Context) error

	// Size returns the number of items in memory for a specific session
	Size(ctx context.Context, sessionID string) (int, error)

	// GetSessions returns all active session IDs
	GetSessions(ctx context.Context) ([]string, error)
}

// GetCriteria defines criteria for retrieving memory items
type GetCriteria struct {
	// Limit limits the number of items to retrieve (0 = no limit)
	Limit int

	// ItemTypes filters by item types (empty = all types)
	ItemTypes []string

	// AgentName filters by agent name (empty = all agents)
	AgentName string

	// Reverse reverses the order of items (newest first if true)
	Reverse bool
}

// InMemoryStorage implements Memory using in-memory storage with session support
type InMemoryStorage struct {
	mu       sync.RWMutex
	sessions map[string][]result.RunItem // sessionID -> items
}

// NewInMemoryStorage creates a new in-memory storage instance
func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		sessions: make(map[string][]result.RunItem),
	}
}

// Add adds a run result to memory for a specific session
func (m *InMemoryStorage) Add(ctx context.Context, sessionID string, runResult *result.RunResult) error {
	if runResult == nil {
		return fmt.Errorf("run result cannot be nil")
	}
	if sessionID == "" {
		return fmt.Errorf("session ID cannot be empty")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Initialize session if it doesn't exist
	if _, exists := m.sessions[sessionID]; !exists {
		m.sessions[sessionID] = make([]result.RunItem, 0)
	}

	// Add all new items from the run result
	for _, item := range runResult.NewItems {
		m.sessions[sessionID] = append(m.sessions[sessionID], item)
	}

	// Also add the final response if it exists
	if runResult.FinalOutput != nil {
		m.sessions[sessionID] = append(m.sessions[sessionID], &result.MessageItem{
			Role:    "assistant",
			Content: runResult.FinalOutput.(string),
		})
	}

	return nil
}

// Get retrieves memory items based on criteria for a specific session
func (m *InMemoryStorage) Get(ctx context.Context, sessionID string, criteria *GetCriteria) ([]result.RunItem, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("session ID cannot be empty")
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	if criteria == nil {
		criteria = &GetCriteria{}
	}

	// Get items for the specific session
	sessionItems, exists := m.sessions[sessionID]
	if !exists {
		return []result.RunItem{}, nil // Return empty slice for non-existent session
	}

	var filtered []result.RunItem

	// Filter by item types if specified
	if len(criteria.ItemTypes) > 0 {
		typeMap := make(map[string]bool)
		for _, t := range criteria.ItemTypes {
			typeMap[t] = true
		}

		for _, item := range sessionItems {
			if typeMap[item.GetType()] {
				filtered = append(filtered, item)
			}
		}
	} else {
		filtered = make([]result.RunItem, len(sessionItems))
		copy(filtered, sessionItems)
	}

	// Filter by agent name if specified
	if criteria.AgentName != "" {
		var agentFiltered []result.RunItem
		for _, item := range filtered {
			if handoffItem, ok := item.(*result.HandoffItem); ok {
				if handoffItem.AgentName == criteria.AgentName {
					agentFiltered = append(agentFiltered, item)
				}
			} else {
				// For non-handoff items, include them if no agent filter
				agentFiltered = append(agentFiltered, item)
			}
		}
		filtered = agentFiltered
	}

	// Reverse order if requested
	if criteria.Reverse {
		for i := len(filtered)/2 - 1; i >= 0; i-- {
			opp := len(filtered) - 1 - i
			filtered[i], filtered[opp] = filtered[opp], filtered[i]
		}
	}

	// Apply limit if specified
	if criteria.Limit > 0 && len(filtered) > criteria.Limit {
		filtered = filtered[:criteria.Limit]
	}

	return filtered, nil
}

// Clear clears all memory for a specific session
func (m *InMemoryStorage) Clear(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return fmt.Errorf("session ID cannot be empty")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.sessions, sessionID)
	return nil
}

// ClearAll clears all memory for all sessions
func (m *InMemoryStorage) ClearAll(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.sessions = make(map[string][]result.RunItem)
	return nil
}

// Size returns the number of items in memory for a specific session
func (m *InMemoryStorage) Size(ctx context.Context, sessionID string) (int, error) {
	if sessionID == "" {
		return 0, fmt.Errorf("session ID cannot be empty")
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	if sessionItems, exists := m.sessions[sessionID]; exists {
		return len(sessionItems), nil
	}
	return 0, nil
}

// GetSessions returns all active session IDs
func (m *InMemoryStorage) GetSessions(ctx context.Context) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sessions := make([]string, 0, len(m.sessions))
	for sessionID := range m.sessions {
		sessions = append(sessions, sessionID)
	}
	return sessions, nil
}

// GetAll returns all items in memory for a specific session
func (m *InMemoryStorage) GetAll(ctx context.Context, sessionID string) ([]result.RunItem, error) {
	return m.Get(ctx, sessionID, &GetCriteria{})
}

// GetRecent returns the most recent items for a specific session
func (m *InMemoryStorage) GetRecent(ctx context.Context, sessionID string, limit int) ([]result.RunItem, error) {
	return m.Get(ctx, sessionID, &GetCriteria{
		Limit:   limit,
		Reverse: true,
	})
}

// GetByType returns items of specific types for a specific session
func (m *InMemoryStorage) GetByType(ctx context.Context, sessionID string, itemTypes []string) ([]result.RunItem, error) {
	return m.Get(ctx, sessionID, &GetCriteria{
		ItemTypes: itemTypes,
	})
}
