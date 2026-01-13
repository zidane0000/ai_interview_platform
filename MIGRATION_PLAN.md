# 🚀 Migration Plan: Two Repos → Monorepo

**Goal:** Combine AI_Interview_Backend + AI_Interview_Frontend into single deployable monorepo

**Timeline:** 2-3 days
**Status:** 🔵 Planning Phase

---

## 📋 Decisions Summary

| Question | Decision | Rationale |
|----------|----------|-----------|
| **Retry Strategy** | Keep frontend retry (reuse existing) | Already implemented, no extra code |
| **Mock API** | Simplify to ~150 lines | Easier maintenance, GitHub Actions testing |
| **Database** | PostgreSQL from day 1 | Already implemented, Railway provides free DB |
| **AI Features** | Keep: Multiple providers<br>Remove: Metrics, caching, health checks | MVP simplicity, document removed features |
| **Deployment** | Railway | Simplest setup, auto-detect Go+Node |

---

## 🎯 Migration Phases

### **Phase 1: Setup Monorepo Structure** ✅ DONE
- [x] Create `/Users/dave_lin/Desktop/personal/Code/ai-interview-platform/` directory
- [x] Create subdirectories: `frontend/`, `.github/workflows/`, `docs/`

---

### **Phase 2: Copy Backend Code** (30 minutes)

**Copy these directories AS-IS:**
```bash
# From AI_Interview_Backend/ to ai-interview-platform/
api/          → api/
data/         → data/
config/       → config/
utils/        → utils/
e2e/          → e2e/
architecture/ → docs/architecture/
main.go       → main.go
go.mod        → go.mod
go.sum        → go.sum
.gitignore    → .gitignore (merge with frontend's)
LICENSE       → LICENSE
```

**Copy ai/ directory WITH modifications:**
```bash
ai/ → ai/  (will simplify in Phase 4)
```

**DON'T copy:**
- `.git/` (will create new repo)
- `.github/` (will create new workflow)
- `README.md` (will write new one at the end)

---

### **Phase 3: Copy Frontend Code** (30 minutes)

**Copy these to frontend/ subdirectory:**
```bash
# From AI_Interview_Frontend/ to ai-interview-platform/frontend/
src/              → frontend/src/
public/           → frontend/public/
index.html        → frontend/index.html
package.json      → frontend/package.json
package-lock.json → frontend/package-lock.json
tsconfig.json     → frontend/tsconfig.json
vite.config.ts    → frontend/vite.config.ts
eslint.config.js  → frontend/eslint.config.js
```

**DON'T copy:**
- `.git/`
- `.github/`
- `README.md`
- `node_modules/` (will reinstall)
- `dist/` (will rebuild)

---

### **Phase 4: Backend Simplification** (2-3 hours)

#### **4.1: Simplify AI Layer**

**Files to modify:**

**DELETE entirely:**
- `ai/enhanced_client.go` (500 lines - metrics, caching, health checks)
- `ai/client_factory.go` (69 lines - factory pattern)

**CREATE new simple client:**
- `ai/client.go` - Direct AI provider calls (~100 lines)

**KEEP unchanged:**
- `ai/openai_provider.go`
- `ai/gemini_provider.go`
- `ai/mock_provider.go`

**New ai/client.go structure:**
```go
type AIClient struct {
    provider Provider  // OpenAI, Gemini, or Mock
}

func NewAIClient(providerType string, apiKeys map[string]string) (*AIClient, error) {
    // Simple provider selection
}

func (c *AIClient) GenerateResponse(ctx context.Context, messages []Message) (string, error) {
    // Direct call to provider, no caching/metrics
}
```

**Update imports in:**
- `api/handlers.go` - Change from `AIClientFactory` to `AIClient`
- `main.go` - Inject single `AIClient` instead of factory

---

#### **4.2: Simplify Data Layer**

**Files to modify:**

**DELETE:**
- `data/hybrid_store.go` (173 lines - if/else branching)
- `data/hybrid_store_test.go`

**KEEP:**
- `data/postgres_store.go` (or rename from db_service.go)
- `data/memory_store.go`

**Update main.go:**
```go
// Before: var GlobalStore *HybridStore
// After: var Store InterviewStore (interface)

// In main():
var store data.InterviewStore
if databaseURL := os.Getenv("DATABASE_URL"); databaseURL != "" {
    store = data.NewPostgresStore(databaseURL)
} else {
    store = data.NewMemoryStore()
}
```

---

#### **4.3: Add Static File Serving**

**Modify main.go:**

Add at top:
```go
import (
    "embed"
    "io/fs"
    "net/http"
)

//go:embed frontend/dist/*
var frontendFS embed.FS
```

Update router setup:
```go
// Prefix all API routes with /api
router.Route("/api", func(r chi.Router) {
    r.Post("/interviews", api.CreateInterviewHandler)
    r.Get("/interviews", api.ListInterviewsHandler)
    // ... all existing routes
})

// Serve frontend static files
frontendDist, _ := fs.Sub(frontendFS, "frontend/dist")
router.Handle("/*", http.FileServer(http.FS(frontendDist)))
```

---

### **Phase 5: Frontend Simplification** (2 hours)

#### **5.1: Simplify Mock API**

**File to modify: `frontend/src/services/mockApi.ts`**

**Before:** 622 lines (12 hardcoded interviews, complex AI simulation)
**After:** ~150 lines (generated data, simple responses)

**New structure:**
```typescript
// Helper to generate mock data
function generateMockInterview(id: number) { ... }
const MOCK_INTERVIEWS = Array.from({length: 12}, (_, i) => generateMockInterview(i));

// Simplified API - just return data with delays
export const mockApi = {
    createInterview: async (data) => { await delay(500); return {...}; },
    getInterviews: async () => { await delay(300); return MOCK_INTERVIEWS; },
    sendMessage: async (sessionId, msg) => {
        await delay(800);
        return { ai_response: "Mock response to: " + msg.message };
    },
    // ... other methods
};
```

**Remove:**
- Complex AI question flow logic (lines 350-500)
- Detailed interview state tracking
- Keep only basic CRUD operations + simple chat

---

#### **5.2: Extract Theme**

**CREATE: `frontend/src/theme/index.ts`**

Move lines 16-134 from App.tsx:
```typescript
import { createTheme } from '@mui/material/styles';

export const appTheme = createTheme({
    palette: { /* ... */ },
    typography: { /* ... */ },
    components: { /* ... */ }
});
```

**UPDATE: `frontend/src/App.tsx`**
```typescript
import { appTheme } from './theme';

function App() {
    return (
        <ThemeProvider theme={appTheme}>
            {/* ... */}
        </ThemeProvider>
    );
}
```

---

#### **5.3: Update API Configuration**

**Modify: `frontend/src/services/api.ts`**

Change base URL:
```typescript
// Before
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';

// After (for production build)
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api';
```

**Update: `frontend/.env.development`**
```bash
VITE_API_BASE_URL=http://localhost:8080/api  # For dev mode (separate servers)
VITE_USE_MOCK_DATA=false
```

**CREATE: `frontend/.env.production`**
```bash
VITE_API_BASE_URL=/api  # Relative path for production
VITE_USE_MOCK_DATA=false
```

---

### **Phase 6: Configuration Files** (30 minutes)

#### **6.1: Railway Configuration**

**CREATE: `railway.toml`**
```toml
[build]
builder = "nixpacks"
buildCommand = "cd frontend && npm ci && npm run build && cd .. && go build -o app"

[deploy]
startCommand = "./app"
restartPolicyType = "on-failure"

[env]
PORT = "8080"
```

---

#### **6.2: Environment Variables**

**CREATE: `.env.example`**
```bash
# Server Configuration
PORT=8080

# Database (PostgreSQL)
DATABASE_URL=postgres://user:password@localhost:5432/ai_interview

# AI Providers (at least one required)
AI_API_KEY=sk-your-openai-api-key-here
GEMINI_API_KEY=your-gemini-api-key-here

# Optional Configuration
SHUTDOWN_TIMEOUT=30s
LOG_LEVEL=info
```

---

#### **6.3: Git Configuration**

**CREATE: `.gitignore` (merge both repos)**
```
# Go
*.exe
*.exe~
*.dll
*.so
*.dylib
*.test
*.out
/vendor/
app

# Frontend
node_modules/
dist/
.env.local
.env.production.local

# Environment
.env

# IDE
.vscode/
.idea/
*.swp
*.swo

# OS
.DS_Store
```

---

### **Phase 7: GitHub Actions** (30 minutes)

**CREATE: `.github/workflows/test.yml`**
```yaml
name: Test

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  backend-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.23'

      - name: Run backend tests
        run: go test ./...

  frontend-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Set up Node.js
        uses: actions/setup-node@v3
        with:
          node-version: '20'

      - name: Install dependencies
        run: cd frontend && npm ci

      - name: Run tests with mock API
        run: cd frontend && npm run test
        env:
          VITE_USE_MOCK_DATA: true

      - name: Build frontend
        run: cd frontend && npm run build
```

---

### **Phase 8: Testing** (1-2 hours)

#### **8.1: Local Testing - Development Mode**

```bash
# Terminal 1: Frontend dev server
cd frontend
npm install
npm run dev  # http://localhost:5173

# Terminal 2: Backend server
go mod download
go run main.go  # http://localhost:8080
```

**Test checklist:**
- [ ] Create interview
- [ ] Start chat session
- [ ] Send messages to AI
- [ ] End session and get evaluation
- [ ] View interview history

---

#### **8.2: Local Testing - Production Build**

```bash
# Build frontend
cd frontend && npm run build

# Build and run single binary
cd .. && go build -o app
./app

# Visit http://localhost:8080
```

**Test checklist:**
- [ ] Frontend loads at http://localhost:8080
- [ ] API calls work (check browser DevTools)
- [ ] All features work as in dev mode

---

#### **8.3: Mock Mode Testing**

```bash
cd frontend
npm run mock  # Run with VITE_USE_MOCK_DATA=true
```

**Test checklist:**
- [ ] Mock indicator shows "Mock Mode"
- [ ] Can create interview
- [ ] Mock AI responds to messages
- [ ] No backend needed

---

### **Phase 9: Documentation** (1 hour)

**CREATE/UPDATE:**
- [x] `MIGRATION_PLAN.md` (this file)
- [ ] `README.md` - Project overview, setup, deployment
- [ ] `docs/REMOVED_FEATURES.md` - Document enterprise features removed
- [ ] `docs/API.md` - API endpoint documentation
- [ ] `docs/DEPLOYMENT.md` - Railway/Zeabur deployment guide

---

### **Phase 10: Deployment** (1 hour)

#### **10.1: Push to GitHub**

```bash
cd /Users/dave_lin/Desktop/personal/Code/ai-interview-platform

git init
git add .
git commit -m "Initial commit: Monorepo migration"

# Create GitHub repo first, then:
git remote add origin https://github.com/YOUR_USERNAME/ai-interview-platform.git
git push -u origin main
```

---

#### **10.2: Deploy to Railway**

1. **Create Railway account**: https://railway.app
2. **Create new project**: "Deploy from GitHub"
3. **Select repository**: ai-interview-platform
4. **Add PostgreSQL**: Click "New" → "Database" → "PostgreSQL"
5. **Set environment variables**:
   - `AI_API_KEY` (your OpenAI key)
   - `GEMINI_API_KEY` (optional)
6. **Deploy**: Railway auto-detects `railway.toml` and builds

**First deployment will:**
- Install Node dependencies
- Build frontend (`npm run build`)
- Build Go binary
- Start server on Railway's PORT
- Auto-connect to PostgreSQL

---

## 📊 Expected Results

### **Before (Current State)**

```
AI_Interview_Backend/     (~2000 lines backend)
  ├── api/
  ├── ai/ (3-layer abstraction, 600+ lines)
  ├── data/ (hybrid store, 173 lines)
  └── ...

AI_Interview_Frontend/    (~3000 lines frontend)
  ├── src/
  │   └── services/
  │       └── mockApi.ts (622 lines)
  └── ...

Total: ~5000 lines, 2 repos, 2 deploys
```

### **After (Monorepo)**

```
ai-interview-platform/    (~3000 lines total)
  ├── frontend/           (~2400 lines)
  │   └── src/
  │       └── services/
  │           └── mockApi.ts (~150 lines)
  ├── api/
  ├── ai/ (simplified, ~150 lines)
  ├── data/ (no hybrid store)
  └── main.go (+ static serving)

Total: ~3000 lines, 1 repo, 1 deploy
Reduction: ~40% fewer lines
```

---

## ✅ Success Criteria

### **Functional:**
- [ ] All interview features work (create, chat, evaluate)
- [ ] PostgreSQL persistence works
- [ ] Multiple AI providers work (OpenAI, Gemini, Mock)
- [ ] Frontend served by Go binary
- [ ] Mock mode works for testing

### **Technical:**
- [ ] Single `git push` deploys everything
- [ ] Frontend build embedded in Go binary
- [ ] All tests pass (backend + frontend)
- [ ] GitHub Actions runs tests on push
- [ ] Railway deployment succeeds

### **Code Quality:**
- [ ] ~1000 lines deleted (40% reduction)
- [ ] No duplicate retry logic
- [ ] Simplified AI layer
- [ ] Clean monorepo structure
- [ ] Updated documentation

---

## 🚨 Rollback Plan

If migration fails:

1. **Original repos still exist** at:
   - `/Users/dave_lin/Desktop/personal/Code/AI_Interview_Backend/`
   - `/Users/dave_lin/Desktop/personal/Code/AI_Interview_Frontend/`

2. **Can revert to separate deployment:**
   - Frontend → Vercel
   - Backend → Railway

3. **Git history preserved:**
   - Original commits in source repos
   - Can cherry-pick fixes back

---

## 📞 Next Steps

**After this plan is approved:**

1. ✅ Phase 2: Copy backend code
2. ✅ Phase 3: Copy frontend code
3. ✅ Phase 4-5: Simplify code
4. ✅ Phase 6-7: Add configs
5. ✅ Phase 8: Test thoroughly
6. ✅ Phase 9: Write docs
7. ✅ Phase 10: Deploy to Railway

**Estimated total time:** 8-12 hours over 2-3 days

---

**Ready to start? Let's begin with Phase 2: Copy backend code!**
