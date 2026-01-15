# AI Interview Platform

A full-stack AI-powered interview platform with Bring Your Own Key (BYOK) support. Conduct intelligent, conversational interviews using your own OpenAI, Gemini, or any OpenAI-compatible API. No vendor lock-in, no hidden costs.

## Key Features

**BYOK Architecture**
- Use your own AI provider API keys (OpenAI, Gemini, Groq, Perplexity, etc.)
- Support for custom OpenAI-compatible endpoints
- Zero platform AI costs - users pay providers directly
- Secure: Keys stored in browser LocalStorage only, never on servers

**Interview Platform**
- Real-time AI conversational interviews
- Multiple interview types: General, Technical, Behavioral
- Detailed AI-powered evaluations with scoring
- Multi-language support: English and Traditional Chinese

**Developer Experience**
- Monorepo: Single repo, single deployment
- Mock AI mode: Develop and demo without API keys
- Production-ready: PaaS compatible (single binary deployment)
- Comprehensive test coverage

## Quick Start

### Prerequisites

- Go 1.23+
- Node.js 20+
- (Optional) PostgreSQL

### Local Development

**1. Clone and setup:**
```bash
git clone https://github.com/YOUR_USERNAME/ai-interview-platform.git
cd ai-interview-platform
```

**2. Start backend:**
```bash
go run main.go
# Runs on http://localhost:8080
# Uses in-memory storage (no database needed)
# Defaults to mock AI (no API keys required)
```

**3. Start frontend (in another terminal):**
```bash
cd frontend
npm install
npm run dev
# Runs on http://localhost:5173
```

**4. Visit http://localhost:5173**
- First-time prompt will ask for API key or skip to demo mode
- Enter your OpenAI/Gemini key to use real AI
- Or click "Try Demo Mode" for mock AI responses

### Production Build

**Build as single binary:**
```bash
# Build frontend
cd frontend && npm install && npm run build

# Build backend (embeds frontend)
cd .. && go build -o app

# Run
./app
# Serves both frontend and API on :8080
```

## BYOK Configuration

### Supported Providers

- **OpenAI** - Default (api.openai.com/v1)
- **Google Gemini** - Google's AI
- **Groq** - Fast inference (api.groq.com/openai/v1)
- **Perplexity** - Search-enhanced AI (api.perplexity.ai)
- **Any OpenAI-compatible API** - Together.ai, local Ollama, etc.

### How BYOK Works

1. **User enters API key** in Settings or first-time prompt
2. **Key stored in browser** (LocalStorage) - never sent to our servers
3. **Every AI request** includes key in HTTPS header
4. **Backend creates ephemeral client** - used once, discarded immediately
5. **User pays provider directly** - no markup, full cost control

**Privacy guarantee:** Backend never logs, stores, or persists user API keys.

### Using Custom Endpoints

In Settings, configure:
- **API Key:** Your provider's key (format varies by provider)
- **Custom Base URL (optional):** e.g., `https://api.groq.com/openai/v1`

Leave base URL empty for standard OpenAI.

## Tech Stack

**Backend:**
- Go 1.23+
- Chi Router (lightweight HTTP routing)
- GORM (PostgreSQL ORM)
- Hybrid storage: Memory (dev) or PostgreSQL (prod)

**Frontend:**
- React 19 with TypeScript
- Material-UI v7
- React Router v7
- Axios with retry logic
- i18next (internationalization)
- Vite (build tool)

**AI Integration:**
- OpenAI API (GPT-4, GPT-3.5)
- Google Gemini API
- OpenAI-compatible endpoints
- Mock provider (development/demo)

## Project Structure

```
ai-interview-platform/
├── api/              # HTTP handlers and routing
├── ai/               # AI provider integrations
├── data/             # Database models and repositories
├── config/           # Configuration management
├── utils/            # Logging and utilities
├── e2e/              # End-to-end tests
├── frontend/         # React frontend application
│   ├── src/
│   │   ├── pages/       # Page components
│   │   ├── components/  # Shared UI components
│   │   ├── services/    # API client and mock
│   │   ├── theme/       # Material-UI theme
│   │   └── types/       # TypeScript definitions
│   └── package.json
├── main.go           # Application entry point
├── go.mod            # Go dependencies
└── README.md
```

## Environment Configuration

**All environment variables are optional.** The app runs with sensible defaults.

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP server port |
| `DATABASE_URL` | *(none)* | PostgreSQL connection (uses memory if not set) |
| `SHUTDOWN_TIMEOUT` | `30s` | Graceful shutdown timeout |

**Note:** With BYOK, you don't need to configure AI provider keys on the server. Users provide their own keys via the UI.

**Frontend environment variables:**
- Automatically configured for monorepo deployment
- See `frontend/.env.development`, `frontend/.env.production` for reference

## API Endpoints

All API routes are prefixed with `/api`:

- `POST /api/interviews` - Create interview
- `GET /api/interviews` - List interviews (with pagination, filtering, sorting)
- `GET /api/interviews/:id` - Get interview details
- `POST /api/interviews/:id/chat/start` - Start AI chat session
- `POST /api/chat/:sessionId/message` - Send message to AI
- `GET /api/chat/:sessionId` - Get chat session
- `POST /api/chat/:sessionId/end` - End session and get evaluation
- `POST /api/evaluation` - Submit traditional evaluation
- `GET /api/evaluation/:id` - Get evaluation results
- `GET /health` - Health check

## Deployment

### PaaS Platforms

The app works with any platform that supports Go + Node.js builds (Render, Fly.io, etc.):

**General deployment steps:**
1. Push code to GitHub
2. Connect your PaaS platform to the repository
3. Configure build and start commands (see below)
4. (Optional) Add PostgreSQL database
5. Deploy

**Build command:**
```bash
cd frontend && npm ci && npm run build && cd .. && go build -o app
```

**Start command:**
```bash
./app
```

**Requirements:**
- Go 1.23+ runtime
- Node.js 20+ for build step
- HTTPS enabled (for secure API key transmission)

**No environment variables required** - BYOK handles AI keys, memory storage works out of the box.

**Optional PostgreSQL:**
Most PaaS platforms provide managed PostgreSQL. Set `DATABASE_URL` environment variable to enable persistent storage.

## Testing

**Backend tests:**
```bash
# All tests
go test ./...

# Unit tests only
go test ./api ./data ./ai ./config ./utils

# E2E tests (requires running server)
go test ./e2e/...
```

**Frontend:**
```bash
cd frontend

# Build test
npm run build

# Lint
npm run lint

# Mock mode test
npm run mock
```

## Development

### Running with Mock AI (No API Keys)

Default mode - no configuration needed:
```bash
go run main.go              # Backend with mock AI
cd frontend && npm run dev  # Frontend
```

All AI responses will be mock/canned responses. Perfect for development and testing.

### Running with Your Own Keys

**Option 1: Via UI (Recommended)**
- Visit Settings page
- Enter your OpenAI/Gemini API key
- Keys stored in your browser only

**Option 2: Via Environment (For Testing)**
```bash
export OPENAI_API_KEY=sk-your-key-here
go run main.go
```

Note: UI-provided keys (BYOK) take precedence over environment variables.

### Database Options

**Memory (Default):**
```bash
go run main.go
# Uses in-memory storage, data resets on restart
```

**PostgreSQL:**
```bash
export DATABASE_URL=postgres://user:password@localhost:5432/ai_interview
go run main.go
# Persistent storage, auto-migrates schema
```

## Architecture Highlights

**BYOK-First Design:**
- No shared AI client
- Ephemeral clients created per-request from user-provided keys
- Falls back to mock if no keys provided
- Backend never stores user API keys

**Hybrid Storage:**
- Adapter pattern switches between Memory and PostgreSQL
- Auto-detects based on DATABASE_URL
- Zero-config development, database-backed production

**Monorepo Deployment:**
- Frontend embedded in Go binary via `//go:embed`
- Single artifact deployment
- Frontend at `/`, API at `/api`

## Security

**API Key Handling:**
- User keys stored in browser LocalStorage only
- Transmitted via HTTPS headers (encrypted)
- Backend creates ephemeral clients (never persisted)
- No server-side key storage, logging, or caching

**CORS:**
- Localhost origins allowed for development
- Configure allowed origins for production

**Timeouts:**
- Read/Write/Idle timeouts configured
- Graceful shutdown with cleanup

## Browser Support

- Chrome 90+
- Firefox 88+
- Safari 14+
- Edge 90+

## License

MIT License - see LICENSE file for details

## Contributing

This is a personal MVP project. Bug reports and suggestions welcome via GitHub Issues.

---

**Built with Go, React, TypeScript, and AI**
