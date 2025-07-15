# Memory Package

The memory package provides conversation history and context management for the Agent SDK. It allows agents to remember previous interactions and maintain context across multiple conversations.

## Features

- **In-Memory Storage**: Fast, thread-safe memory storage for the current session
- **Flexible Retrieval**: Query memory by type, agent, with limits, and ordering options
- **Automatic Integration**: Seamless integration with the Runner for automatic memory management
- **Thread-Safe**: All operations are protected with proper locking mechanisms

## Usage

### Basic Usage

```go
package main

import (
    "context"
    "github.com/pontus-devoteam/agent-sdk-go/pkg/memory"
    "github.com/pontus-devoteam/agent-sdk-go/pkg/runner"
)

func main() {
    // Create memory storage
    memoryStorage := memory.NewInMemoryStorage()
    
    // Create runner with memory
    r := runner.NewRunner()
    r.WithMemory(memoryStorage)
    
    // Now all agent runs will automatically use memory
    result, err := r.RunSync(agent, &runner.RunOptions{
        Input: "Hello, remember my name is Alice",
    })
}
```

### Manual Memory Management

```go
// Create memory storage
memoryStorage := memory.NewInMemoryStorage()
ctx := context.Background()

// Add run result to memory
err := memoryStorage.Add(ctx, runResult)

// Get all memory items
allItems, err := memoryStorage.GetAll(ctx)

// Get recent items
recentItems, err := memoryStorage.GetRecent(ctx, 5)

// Get specific item types
messageItems, err := memoryStorage.GetByType(ctx, []string{"message"})

// Get with custom criteria
items, err := memoryStorage.Get(ctx, &memory.GetCriteria{
    Limit:     10,
    ItemTypes: []string{"message", "tool_call"},
    Reverse:   true, // newest first
})

// Clear memory
err := memoryStorage.Clear(ctx)

// Get memory size
size, err := memoryStorage.Size(ctx)
```

## Memory Interface

The `Memory` interface defines the contract for memory implementations:

```go
type Memory interface {
    // Add adds a run result to memory
    Add(ctx context.Context, runResult *result.RunResult) error
    
    // Get retrieves memory items based on criteria
    Get(ctx context.Context, criteria *GetCriteria) ([]result.RunItem, error)
    
    // Clear clears all memory
    Clear(ctx context.Context) error
    
    // Size returns the number of items in memory
    Size(ctx context.Context) (int, error)
}
```

## GetCriteria Options

```go
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
```

## Item Types

The memory system stores different types of items:

- **message**: User and assistant messages
- **tool_call**: Function/tool calls made by agents
- **tool_result**: Results from tool executions
- **handoff**: Agent handoff operations

## Integration with Runner

When memory is configured with a Runner, it automatically:

1. **Loads Context**: Prepends memory items to the input for each agent run
2. **Stores Results**: Saves new items from each run result to memory
3. **Maintains History**: Preserves conversation flow across multiple interactions

## Example: Memory-Enabled Chat Agent

See `examples/memory_example/main.go` for a complete example of using memory with a conversational agent that can remember previous interactions.

## Thread Safety

All memory operations are thread-safe and can be used concurrently from multiple goroutines. The `InMemoryStorage` implementation uses read-write mutexes to ensure safe concurrent access.

## Future Extensions

The memory interface is designed to support future implementations such as:

- File-based persistent storage
- Database-backed memory
- Distributed memory systems
- Memory with expiration policies
- Encrypted memory storage
