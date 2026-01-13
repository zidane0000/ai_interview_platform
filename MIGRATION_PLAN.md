#  Migration Plan: Two Repos → Monorepo

**Goal:** Combine AI_Interview_Backend + AI_Interview_Frontend into single deployable monorepo

**Timeline:** 2-3 days
**Status:**  In Progress (Phase 4.1 Complete)

---

##  Decisions Summary

| Question | Decision | Rationale |
|----------|----------|-----------|
| **Retry Strategy** | Keep frontend retry (reuse existing) | Already implemented, no extra code |
| **Mock API** | Keep as-is (622 lines) REVISED | Valuable dev tool for frontend-only development |
| **Database** | PostgreSQL from day 1 | Already implemented, Railway provides free DB |
| **AI Features** | Keep: Multiple providers (OpenAI, Gemini, Mock)<br>Remove: Metrics, caching, factory pattern  DONE | MVP simplicity, ~692 lines removed |
| **Data Layer** | Keep hybrid_store.go (Adapter pattern)  REVISED | Good architecture for MVP flexibility |
| **Deployment** | Railway | Simplest setup, auto-detect Go+Node |

---

##  Migration Phases

### **Phase 1: Setup Monorepo Structure**  DONE
- [x] Create `/Users/dave_lin/Desktop/personal/Code/ai-interview-platform/` directory
- [x] Create subdirectories: `frontend/`, `.github/workflows/`, `docs/`

---

### **Phase 2: Copy Backend Code**  DONE

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

### **Phase 3: Copy Frontend Code**  DONE

**CRITICAL FIXES APPLIED:**
- Added missing .env files (.env.development, .env.mock, .env.production)
- Added missing tsconfig files (tsconfig.app.json, tsconfig.node.json)
- Updated go.mod module path to github.com/zidane0000/ai-interview-platform
- Updated 27 import statements in .go files
- Updated .gitignore for monorepo

---

### **Phase 3: Copy Frontend Files**  DONE

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

### **Phase 4: Backend Simplification**

#### **4.1: Simplify AI Layer**  COMPLETED

**Files to modify:**

**DELETE entirely:**
- `ai/enhanced_client.go` (500 lines - metrics, caching, health checks)
- `ai/client_factory.go` (69 lines - factory pattern)

**CREATE new simple client:**
- `ai/client.go` - Direct AI provider calls (~210 lines)  COMPLETED

**KEEP unchanged:**
- `ai/openai_provider.go`
- `ai/gemini_provider.go`
- `ai/mock_provider.go`

**New ai/client.go structure:**
```go
type AIClient struct {
    provider AIProvider  // OpenAI, Gemini, or Mock
    config   *AIConfig
}

func NewAIClient(cfg *AIConfig) (*AIClient, error) {
    // Creates provider based on cfg.DefaultProvider
    // No factory pattern, no metrics, no caching
}

func (c *AIClient) GenerateChatResponseWithLanguage(...) (string, error) {
    // Direct call to provider.GenerateResponse()
}

func (c *AIClient) EvaluateAnswersWithContext(...) (float64, string, error) {
    // Direct call to provider.EvaluateAnswers()
}
```

**Update imports in:**
- `api/handlers.go` - Change from `AIClientFactory` to `AIClient`
- `main.go` - Inject single `AIClient` instead of factory

---

#### **4.2: Clean Up Data Layer** COMPLETED

**DECISION: Keep hybrid_store.go - it's good architecture!**

After comprehensive analysis, hybrid_store.go is an **Adapter Pattern** that:
- Routes calls between MemoryStore and DatabaseService (different interfaces)
- Provides auto-detection via DATABASE_URL (zero-config deployment)
- Converts between different query formats (ListInterviewsOptions ↔ InterviewFilters)
- The 14 if/else methods are necessary routing logic, NOT duplication

**Changes made (Commit 8af7c95):**
- Removed legacy unused variable: `var Store = NewMemoryStore()`
- Added comprehensive documentation to hybrid_store.go explaining Adapter Pattern

**Files kept unchanged:**
- data/hybrid_store.go (211 lines) - Adapter pattern, well-designed
- data/memory_store.go (now 250 lines, -1 line)
- data/db_service.go (62 lines) - Database coordinator
- All repository files - Clean data access layer

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

### **Phase 5: Frontend Refinement** REVISED

#### **5.1: Keep Mock API** REVISED - NO CHANGES NEEDED

**DECISION: Keep mockApi.ts as-is (622 lines)**

After comprehensive analysis, mockApi.ts is **NOT over-engineered**:
- Provides complete mock backend for frontend-only development
- Language-aware AI responses (EN/ZH-TW) with 8 progressive questions
- Realistic data (12 diverse interviews, international names)
- Proper pagination, filtering, sorting simulation
- Chat session state management
- Enables GitHub Actions testing without backend
- Critical value for solo developer workflow

**Original plan was WRONG** - assumed it was duplication, but it's actually a valuable development tool.

**NO ACTION NEEDED** - Keep all 622 lines

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

#### **5.3: Update API Configuration** PARTIALLY DONE

**STATUS:**
- .env.development - DONE in Phase 4.3 (updated to http://localhost:8080/api)
- .env.production - DONE in Phase 4.3 (updated to /api)
- api.ts default URL - NOT DONE YET

**Remaining change: `frontend/src/services/api.ts` line 17**

Update default URL for production builds:
```typescript
// Before
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';

// After (fallback to /api for production)
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api';
```

This ensures production builds work even if VITE_API_BASE_URL is not set.

---

#### **5.4: Remove Unused Files** NEW

**Check if CreateInterview.tsx is used:**
- Home.tsx handles interview creation
- CreateInterview.tsx (274 lines) might be unused duplicate
- Search codebase for imports/references

**If unused:**
- Delete frontend/src/pages/CreateInterview.tsx
- Remove from App.tsx routes (if present)
- Impact: -274 lines

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

#### **6.3: Git Configuration** DONE (Phase 3)

**SKIP - .gitignore already completed in Phase 3:**
- Copied from AI_Interview_Backend
- Updated with `app` binary
- Updated with frontend ignores (node_modules/, frontend/dist/)
- Updated to commit .env.development, .env.mock, .env.production (templates)

No action needed.

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

#### **10.1: Push to GitHub**  REVISED

```bash
cd /Users/dave_lin/Desktop/personal/Code/ai-interview-platform

# Git already initialized with commits:
# - b1d2dc8: Initial monorepo setup
# - 9cf1bb8: Simplified AI layer

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

##  Expected Results

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

### **After (Monorepo)**  REVISED

```
ai-interview-platform/    (~4800 lines total)
  ├── frontend/           (~3000 lines - kept mockApi)
  │   └── src/
  │       └── services/
  │           └── mockApi.ts (622 lines - KEPT, valuable dev tool)
  ├── api/                (~850 lines)
  ├── ai/                 (~2155 lines - removed 692 lines)
  │   ├── client.go       (210 lines - simplified)
  │   ├── openai_provider.go (477 lines - kept)
  │   ├── gemini_provider.go (540 lines - kept)
  │   └── mock_provider.go (129 lines - kept)
  ├── data/               (~900 lines - kept hybrid_store)
  │   └── hybrid_store.go (211 lines - KEPT, Adapter pattern)
  └── main.go             (+ static serving)

Total: ~4800 lines, 1 repo, 1 deploy
Reduction: ~700 lines (AI metrics/caching only, kept mockAPI)
Percentage: ~12-15% reduction

Note: Kept mockApi.ts - provides huge value for solo developer workflow
```

---

##  Success Criteria

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
- [ ] ~700 lines deleted (12-15% reduction) REVISED AGAIN
- [ ] Simplified AI layer (removed metrics/caching, kept providers)
- [ ] Kept hybrid_store.go (Adapter pattern) and mockApi.ts (dev tool)
- [ ] Clean monorepo structure
- [ ] Updated documentation

---

##  Rollback Plan

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

##  Next Steps

**After this plan is approved:**

1.  Phase 2: Copy backend code
2.  Phase 3: Copy frontend code
3.  Phase 4-5: Simplify code
4.  Phase 6-7: Add configs
5.  Phase 8: Test thoroughly
6.  Phase 9: Write docs
7.  Phase 10: Deploy to Railway

**Estimated total time:** 8-12 hours over 2-3 days

---

**Ready to start? Let's begin with Phase 2: Copy backend code!**
