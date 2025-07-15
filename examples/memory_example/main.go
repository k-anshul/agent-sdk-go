package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/pontus-devoteam/agent-sdk-go/pkg/agent"
	"github.com/pontus-devoteam/agent-sdk-go/pkg/memory"
	"github.com/pontus-devoteam/agent-sdk-go/pkg/model/providers/openai"
	"github.com/pontus-devoteam/agent-sdk-go/pkg/result"
	"github.com/pontus-devoteam/agent-sdk-go/pkg/runner"
	"github.com/pontus-devoteam/agent-sdk-go/pkg/tool"
)

func main() {
	// Create an OpenAI provider (you'll need to set your API key)
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	provider := openai.NewProvider(apiKey)
	provider.SetDefaultModel("gpt-4o")
	provider.WithRateLimit(40, 80000)
	provider.WithRetryConfig(3, 2*time.Second)

	// Create memory storage
	memoryStorage := memory.NewInMemoryStorage()

	// Create a simple tool
	getTimeInfo := tool.NewFunctionTool(
		"get_time_info",
		"Get current time information",
		func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{
				"current_time": time.Now().Format(time.RFC3339),
				"timezone":     "UTC",
			}, nil
		},
	)

	// Create an agent
	chatAgent := agent.NewAgent("ChatBot")
	chatAgent.SetModelProvider(provider)
	chatAgent.WithModel("gpt-4o")
	chatAgent.WithTools(getTimeInfo)
	chatAgent.SetSystemInstructions(`You are a helpful assistant with memory. 
You can remember previous conversations and refer back to them. 
You also have access to current time information when needed.
Be conversational and remember what the user has told you before.`)

	// Create runner with memory
	r := runner.NewRunner()
	r.WithDefaultProvider(provider)
	r.WithMemory(memoryStorage)

	fmt.Println("Memory-enabled Chat Agent Example")
	fmt.Println("================================")

	// Simulate a multi-turn conversation
	conversations := []string{
		"Hi, my name is Alice and I'm a software engineer.",
		"What's my name?",
		"What time is it?",
		"What do I do for work?",
		"Can you summarize our conversation so far?",
	}

	for i, userInput := range conversations {
		fmt.Printf("\n--- Turn %d ---\n", i+1)
		fmt.Printf("User: %s\n", userInput)

		// Run the agent
		result, err := r.RunSync(chatAgent, &runner.RunOptions{
			Input:     userInput,
			SessionID: "alice-session", // Use a consistent session ID
			MaxTurns:  3,
		})

		if err != nil {
			log.Printf("Error: %v", err)
			continue
		}

		fmt.Printf("Assistant: %s\n", result.FinalOutput)

		// Show memory size after each interaction
		ctx := context.Background()
		size, err := memoryStorage.Size(ctx, "alice-session")
		if err == nil {
			fmt.Printf("Memory size: %d items\n", size)
		}
	}

	// Demonstrate memory retrieval
	fmt.Println("\n--- Memory Analysis ---")
	ctx := context.Background()
	sessionID := "alice-session"

	// Get all memory items
	allItems, err := memoryStorage.GetAll(ctx, sessionID)
	if err != nil {
		log.Printf("Error getting memory items: %v", err)
		return
	}

	fmt.Printf("Total memory items: %d\n", len(allItems))

	// Get only message items
	messageItems, err := memoryStorage.GetByType(ctx, sessionID, []string{"message"})
	if err != nil {
		log.Printf("Error getting message items: %v", err)
		return
	}

	fmt.Printf("Message items: %d\n", len(messageItems))
	fmt.Println("\nConversation history:")
	for i, item := range messageItems {
		if msgItem, ok := item.(*result.MessageItem); ok {
			fmt.Printf("%d. [%s]: %s\n", i+1, msgItem.Role, msgItem.Content)
		}
	}

	// Get recent items
	recentItems, err := memoryStorage.GetRecent(ctx, sessionID, 3)
	if err != nil {
		log.Printf("Error getting recent items: %v", err)
		return
	}

	fmt.Printf("\nMost recent 3 items:\n")
	for i, item := range recentItems {
		fmt.Printf("%d. Type: %s\n", i+1, item.GetType())
	}

	// Clear memory
	fmt.Println("\nClearing memory...")
	err = memoryStorage.Clear(ctx, sessionID)
	if err != nil {
		log.Printf("Error clearing memory: %v", err)
		return
	}

	size, _ := memoryStorage.Size(ctx, sessionID)
	fmt.Printf("Memory size after clear: %d items\n", size)

	fmt.Println("\nMemory example completed!")
}
