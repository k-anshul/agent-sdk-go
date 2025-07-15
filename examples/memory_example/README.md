# Memory Example

This example demonstrates how to use the memory functionality in the Agent SDK to create agents that can remember previous conversations and maintain context across multiple interactions.

## Features Demonstrated

- **Memory Integration**: Using `InMemoryStorage` with the `Runner`
- **Context Preservation**: Agent remembers user information across turns
- **Multi-turn Conversations**: Simulating a realistic chat scenario
- **Memory Querying**: Retrieving and analyzing stored memory items
- **Memory Management**: Clearing memory when needed

## Setup

1. Set your OpenAI API key in the code:
   ```go
   provider := openai.NewProvider("your-api-key-here")
   ```

2. Run the example:
   ```bash
   go run main.go
   ```

## What the Example Does

1. **Creates a Memory-Enabled Agent**: Sets up an agent with memory storage
2. **Simulates Conversations**: Runs through multiple conversation turns
3. **Demonstrates Memory**: Shows how the agent remembers previous information
4. **Analyzes Memory**: Displays memory contents and statistics
5. **Cleans Up**: Shows how to clear memory

## Sample Output

The example will show a conversation like:

```
--- Turn 1 ---
User: Hi, my name is Alice and I'm a software engineer.
Assistant: Hello Alice! Nice to meet you. It's great to know you're a software engineer...

--- Turn 2 ---
User: What's my name?
Assistant: Your name is Alice!

--- Turn 3 ---
User: What do I do for work?
Assistant: You're a software engineer!
```

## Key Components

- **Memory Storage**: `memory.NewInMemoryStorage()`
- **Runner Integration**: `r.WithMemory(memoryStorage)`
- **Memory Analysis**: Querying memory by type and recency
- **Automatic Context**: Memory items are automatically included in agent context

## Notes

- Replace the API key placeholder with your actual OpenAI API key
- The agent will remember information across all conversation turns
- Memory persists for the duration of the program execution
- You can extend this example to use persistent storage in the future
