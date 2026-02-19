# Vayuu Memory System - Technical Documentation

## Overview

Vayuu has a sophisticated multi-layered memory system that enables the AI agent to learn about the user over time. The memory system is designed to reduce context size when calling the LLM while still providing relevant context.

---

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        Agent                                  │
│  - Receives user messages                                    │
│  - Queries memory before responding                           │
│  - Stores conversation results                                │
└──────────────────────┬──────────────────────────────────────┘
                       │
        ┌──────────────┴──────────────┐
        ▼                                 ▼
┌───────────────────┐     ┌─────────────────────────────┐
│  File Storage     │     │     Memory Manager           │
│  (JSONL logs)     │     │  ┌─────────────────────────┐│
│                   │     │  │  Vector Store (Qdrant) ││
│  - Daily logs     │     │  │  - Semantic search      ││
│  - Full history   │     │  │  - Embeddings          ││
└───────────────────┘     │  └─────────────────────────┘│
                          │  ┌─────────────────────────┐│
                          │  │  SQLite Database       ││
                          │  │  - User profile       ││
                          │  │  - Preferences         ││
                          │  │  - Topics             ││
                          │  └─────────────────────────┘│
                          │  ┌─────────────────────────┐│
                          │  │  Fact Extractor (LLM)   ││
                          │  │  - Extracts structured  ││
                          │  │    facts from conv     ││
                          │  └─────────────────────────┘│
                          └───────────────────────────────┘
```

---

## Memory Layers

### 1. File Storage (Episodic Memory)

**File**: `internal/agent/memory.go`

Stores full conversation history in daily JSONL files.

- **Location**: `{workDir}/memory/YYYY-MM-DD.jsonl`
- **Format**: One JSON object per line:
  ```json
  {"timestamp": "15:30:45", "role": "user", "content": "Hello"}
  {"timestamp": "15:30:46", "role": "assistant", "content": "Hi there!"}
  ```
- **Rotation**: Archives to `YYYY-MM-DD.jsonl.{timestamp}` when file exceeds 10MB

**Purpose**: Complete audit trail, recovery, full conversation context when needed.

---

### 2. Vector Store (Semantic Memory)

**File**: `internal/memory/store.go`

Stores embeddings for semantic search.

- **Backend**: Qdrant (vector database) via REST API
- **Collection**: `vayuu_memory`
- **Vector Dimension**: 768 (for nomic-embed-text)
- **Distance Metric**: Cosine similarity

**Data Stored**:
```json
{
  "id": "uuid",
  "vector": [...],
  "payload": {
    "content": "The user prefers dark mode",
    "type": "fact|preference|knowledge",
    "created_at": "2026-02-19T15:30:45Z"
  }
}
```

**Purpose**: Semantic search - find relevant memories by meaning, not just keywords.

---

### 3. SQLite Database (Structured Memory)

**File**: `internal/memory/database.go`

Structured, queryable data.

**Tables**:

#### `user_profile`
Key-value store for user facts.
```sql
CREATE TABLE user_profile (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at TEXT NOT NULL
);
```
Examples: `name: "John"`, `bio: "Software developer"`, `location: "Bangalore"`

#### `preferences`
User preferences with confidence scoring.
```sql
CREATE TABLE preferences (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    key TEXT NOT NULL,
    value TEXT NOT NULL,
    category TEXT NOT NULL,
    confidence REAL DEFAULT 1.0,
    updated_at TEXT NOT NULL
);
```
- **Confidence**: Increases by 0.1 on each mention (max 1.0)
- **Categories**: `food`, `hobby`, `work`, `communication`, `other`

#### `topics`
Track conversation topics.
```sql
CREATE TABLE topics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    mentions INTEGER DEFAULT 1,
    last_mentioned TEXT NOT NULL
);
```

---

### 4. Embedder

**File**: `internal/memory/embedder.go`

Generates embeddings using Ollama.

- **URL**: `http://localhost:11434` (configurable)
- **Model**: `nomic-embed-text` (configurable)
- **API Endpoint**: `/api/embeddings`

**Purpose**: Convert text to vectors for semantic search.

---

### 5. Fact Extractor

**File**: `internal/memory/extractor.go`

Uses LLM to extract structured facts from conversations.

**Extraction Prompt**:
```
Analyze the following conversation and extract structured information about the user.

Extract facts in this JSON format:
[
  {"type": "fact", "key": "key_name", "value": "the fact"},
  {"type": "preference", "key": "preference_name", "value": "the preference", "category": "category"},
  {"type": "topic", "key": "topic_name", "value": "topic"}
]

Rules:
- Only extract if there's clear new information about the user
- "key" should be lowercase snake_case
- "category" for preferences: food, hobby, work, communication, other
- Return empty array if nothing significant to extract
- Keep values concise (under 50 words)
```

**Output**:
```go
type ExtractedFact struct {
    Type     string // "fact", "preference", "topic"
    Key      string // e.g., "name", "food", "hobby"
    Value    string // e.g., "John", "sushi", "photography"
    Category string // for preferences: food, hobby, work, communication, other
}
```

---

## Memory Manager

**File**: `internal/memory/manager.go`

Orchestrates all memory components.

### Key Functions:

#### `AddMemory(ctx, content, type, metadata)`
- Generates embedding via Ollama
- Stores in Qdrant with metadata
- Increments memory count

#### `SearchMemory(ctx, query, limit)`
- Generates embedding for query
- Searches Qdrant for similar vectors
- Returns ranked results with scores

#### `GetContext(ctx, query, maxTokens)`
- Searches vector store for relevant memories
- Gets user profile from SQLite
- Combines into context string
- Limits to maxTokens (~4 chars per token)

#### `ProcessConversation(ctx, userInput, assistantResponse)`
- Calls Fact Extractor with conversation
- Stores extracted facts:
  - `fact` → vector store + `user_profile` table
  - `preference` → vector store + `preferences` table
  - `topic` → vector store + `topics` table + increment count

---

## Agent Integration

**File**: `internal/agent/agent.go`

### Initialization (`CreateAgent`)
```go
// Create memory manager with database
mgr, err := memory.NewMemoryManagerWithDB(cfg.AgentWorkDir, cfg)
```

### On User Message (`RunAgent`)
1. Get user input
2. Query memory manager for relevant context:
   ```go
   memContext, err := m.memoryMgr.GetContext(ctx, userInput, 500)
   ```
3. Append context to system prompt
4. Call LLM with expanded context

### After Assistant Response
1. Persist full conversation to file storage
2. Async: Process conversation for facts:
   ```go
   go func() {
       m.memoryMgr.ProcessConversation(ctx, userInput, response)
   }()
   ```

---

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `TELEGRAM_TOKEN` | Bot token from BotFather | Required |
| `API_KEY` | LLM API key | Required |
| `API_BASE_URL` | LLM API base URL | `http://localhost:11434/v1` |
| `MODEL` | Model name | `kimi-k2.5:cloud` |
| `AGENT_WORKDIR` | Working directory | `~/.vayuu/workspace` |
| `ALLOWED_USERNAME` | Allowed Telegram user | Required |
| `OLLAMA_BASE_URL` | Ollama server URL | `http://localhost:11434` |
| `OLLAMA_MODEL` | Embedding model | `nomic-embed-text` |

---

## External Services Required

### 1. Ollama (Embeddings)

```bash
# Start server
ollama serve

# Pull embedding model
ollama pull nomic-embed-text
```

**Purpose**: Generate text embeddings for semantic search.

### 2. Qdrant (Vector Database)

```bash
# Run via Docker
docker run -p 6333:6333 -p 6334:6334 qdrant/qdrant
```

**Purpose**: Store and search vector embeddings.

---

## Data Flow Example

### Scenario: User says "I love sushi and Tokyo"

#### Step 1: Query Context
```
User Input: "What's a good restaurant?"
Memory Search: "sushi" → finds "user loves sushi"
User Profile: name="John", food="sushi"
```

**Context Provided to LLM**:
```
Relevant memories:
[preference] The user loves sushi

User Profile:
- name: John
- food: sushi

User: What's a good restaurant?
```

#### Step 2: Assistant Responds
```
Assistant: "Since you love sushi, I recommend..."
```

#### Step 3: Process Conversation (Async)
- LLM extracts: `[{"type": "preference", "key": "food", "value": "sushi", "category": "food"}]`
- Stored in:
  - **Vector DB**: `"preference: user loves sushi"`
  - **SQLite**: `preferences` table updated with `food: sushi`
- User profile updated (if applicable)

---

## Token Management

### Context Budget
- **Max tokens for memory**: 500 (~2000 chars)
- **Priority**:
  1. User profile (SQLite summary)
  2. Recent vector search results

### Search Parameters
- **Limit**: 10 results from vector search
- **Filter**: None (all types)
- **Score Threshold**: None (return all, let LLM decide relevance)

---

## Error Handling

### Memory Manager Unavailable
If Qdrant/Ollama fails:
- Agent continues without memory
- Logs warning
- No crash

### Fact Extraction Fails
- Silently ignored
- No crash
- Continues to next fact

### Database Unavailable
- Vector-only mode activated
- User profile not available in context

---

## Future Improvements

1. **Memory Consolidation**: Merge duplicate/overlapping memories
2. **Importance Scoring**: Not all memories are equal
3. **Forgetting**: Prune low-value memories
4. **Graph Relationships**: Track topic → conversation → facts
5. **Time-Based Decay**: Older memories weighted less

---

## File Structure

```
internal/memory/
├── types.go      # Data types: Memory, SearchResult, Config
├── embedder.go   # Ollama client for embeddings
├── store.go     # Qdrant REST API client
├── database.go   # SQLite for structured data
├── extractor.go # LLM-based fact extraction
└── manager.go   # Orchestrates all components
```

---

## Testing

To verify memory system works:
1. Run Ollama: `ollama serve && ollama pull nomic-embed-text`
2. Run Qdrant: `docker run -p 6333:6333 qdrant/qdrant`
3. Run Vayuu: `go run ./cmd/vayuu`
4. Send message to bot: "My name is [Your Name] and I love pizza"
5. Wait a moment, then ask: "What do you know about me?"

---

## Troubleshooting

### "failed to initialize memory manager"
- Check Qdrant is running on port 6333
- Check Ollama is running on port 11434

### "embedding generation failed"
- Verify `ollama pull nomic-embed-text` completed
- Check Ollama logs

### "search memory failed"
- Check Qdrant collection exists
- Verify vectors were stored

### "parse facts" error
- LLM returned invalid JSON
- Check extractor prompt

---

## Summary

The memory system provides:
- ✅ **Semantic Search**: Find relevant past context by meaning
- ✅ **Structured Data**: Queryable user profile and preferences
- ✅ **Auto-Learning**: LLM extracts facts automatically
- ✅ **Context Reduction**: Only relevant memories sent to LLM
- ✅ **Fault Tolerance**: Works without external services
- ✅ **Privacy**: All data stored locally
