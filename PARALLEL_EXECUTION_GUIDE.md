# Parallel Execution Guide for Agent Payment MCP Implementation

This guide identifies all tasks in the IMPLEMENT_PLAN.md that can be executed asynchronously using subagents for maximum efficiency.

---

## Execution Strategy

### Phase 1: Foundation (Sequential)
1. Create project structure directories
2. Initialize Vite project
3. Initialize Go module

### Phase 2: PWA Library Files (PARALLEL - 3 agents)
**Run simultaneously:** crypto.ts, store.ts, api.ts

### Phase 3: PWA Components (PARALLEL - 4 agents)
**Run simultaneously:** Header.tsx, Settings.tsx, Tools.tsx, Install.tsx

### Phase 4: PWA Main Files (PARALLEL - 3 agents)
**Run simultaneously:** App.tsx + main.tsx, styles.css, PWA essentials (manifest + sw.js + index.html)

### Phase 5: Installation Scripts (PARALLEL - 3 agents)
**Run simultaneously:** install-macos.sh, install-windows.ps1, install-linux.sh

### Phase 6: Build Scripts (PARALLEL - 3 agents)
**Run simultaneously:** build-all.sh, package-mcpb.sh, package-installers.sh

### Phase 7: CI/CD & Docs (PARALLEL - 2 agents)
**Run simultaneously:** GitHub Actions workflow, Documentation files

---

## Subagent Prompts

### PHASE 2: PWA Library Files

#### Agent 2.1: crypto.ts
```
Create the file `/home/richard/Documents/agentpmt/local_mcp/pwa/src/lib/crypto.ts` exactly as specified in section 1.3 of IMPLEMENT_PLAN.md.

File content:
- WebCrypto utilities for AES-GCM encryption
- Functions: generateKey, importKey, exportKey, encryptJSON, decryptJSON
- TypeScript with proper types
- Comments explaining each function

Copy the exact code from IMPLEMENT_PLAN.md lines 356-412.
```

#### Agent 2.2: store.ts
```
Create the file `/home/richard/Documents/agentpmt/local_mcp/pwa/src/lib/store.ts` exactly as specified in section 1.3 of IMPLEMENT_PLAN.md.

File content:
- IndexedDB storage wrapper using 'idb' package
- Encrypted secret storage using crypto.ts functions
- Functions: saveSecrets, loadSecrets, clearSecrets
- TypeScript with proper interfaces (SecretBundle, AgentPayDB)
- localStorage integration for symmetric key storage

Copy the exact code from IMPLEMENT_PLAN.md lines 414-525.
```

#### Agent 2.3: api.ts
```
Create the file `/home/richard/Documents/agentpmt/local_mcp/pwa/src/lib/api.ts` exactly as specified in section 1.4 of IMPLEMENT_PLAN.md.

File content:
- REST API client for Agent Payment endpoints
- Functions: fetchTools, purchaseTool
- TypeScript interfaces: ToolParameter, ToolFunction, ToolRecord, FetchToolsResponse, PurchaseToolResponse, ApiCredentials
- Base URL: https://api.agentpmt.com
- Proper header handling for x-api-key, x-budget-key, Authorization

Copy the exact code from IMPLEMENT_PLAN.md lines 527-633.
```

---

### PHASE 3: PWA Components

#### Agent 3.1: Header.tsx
```
Create the file `/home/richard/Documents/agentpmt/local_mcp/pwa/src/components/Header.tsx` exactly as specified in section 1.5 of IMPLEMENT_PLAN.md.

File content:
- React component for app header
- Logo display with proper image path
- Title "agent PAYMENT"
- Minimal styling

Copy the exact code from IMPLEMENT_PLAN.md (section 1.5, Header component).
```

#### Agent 3.2: Settings.tsx
```
Create the file `/home/richard/Documents/agentpmt/local_mcp/pwa/src/routes/Settings.tsx` exactly as specified in section 1.5 of IMPLEMENT_PLAN.md.

File content:
- React component for API key entry
- Form fields: apiKey, budgetKey, auth (optional)
- Integration with store.ts (loadSecrets, saveSecrets, clearSecrets)
- Save and Clear buttons
- Status messages
- useEffect for loading existing credentials
- TypeScript with proper types

Copy the exact code from IMPLEMENT_PLAN.md (section 1.5, Settings component, approximately lines 636-755).
```

#### Agent 3.3: Tools.tsx
```
Create the file `/home/richard/Documents/agentpmt/local_mcp/pwa/src/routes/Tools.tsx` exactly as specified in section 1.5 of IMPLEMENT_PLAN.md.

File content:
- React component for browsing and testing tools
- ToolCard subcomponent for individual tools
- Integration with api.ts (fetchTools, purchaseTool)
- Integration with store.ts (loadSecrets)
- Dynamic parameter form generation from tool schemas
- Result display
- Loading and error states
- TypeScript with proper types

Copy the exact code from IMPLEMENT_PLAN.md (section 1.5, Tools component, approximately lines 757-897).
```

#### Agent 3.4: Install.tsx
```
Create the file `/home/richard/Documents/agentpmt/local_mcp/pwa/src/routes/Install.tsx` exactly as specified in section 1.5 of IMPLEMENT_PLAN.md.

File content:
- React component for downloading installers
- Editor selection: Claude Desktop, Cursor, VS Code
- Platform selection: Windows, macOS (Intel/ARM), Linux
- Install method selection (for Claude): .mcpb vs script
- Download functions: downloadMcpbPackage, downloadScriptPackage
- Helper functions: getBinaryName, getBinaryPath, generateInstallScript, generateReadme, downloadBlob
- Integration with store.ts and JSZip
- Installation instructions for each method
- TypeScript with proper types

This is a LARGE file. Copy the exact code from IMPLEMENT_PLAN.md (section 1.5, Install component, approximately lines 899-1502).
```

---

### PHASE 4: PWA Main Files

#### Agent 4.1: App.tsx + main.tsx
```
Create TWO files:

1. `/home/richard/Documents/agentpmt/local_mcp/pwa/src/App.tsx`
- Main application component with tab-based routing
- Tab navigation: Settings, Tools, Install
- Header and footer components
- Tab state management
- TypeScript

2. `/home/richard/Documents/agentpmt/local_mcp/pwa/src/main.tsx`
- React application entry point
- ReactDOM.createRoot
- Imports App and styles.css
- StrictMode wrapper

Copy the exact code from IMPLEMENT_PLAN.md section 1.6 (approximately lines 1504-1581).
```

#### Agent 4.2: styles.css
```
Create the file `/home/richard/Documents/agentpmt/local_mcp/pwa/src/styles.css` exactly as specified in section 1.6 of IMPLEMENT_PLAN.md.

File content:
- Global CSS styles for the PWA
- CSS custom properties for colors (light/dark mode)
- Styles for: app layout, header, tabs, content, cards, forms, buttons, status messages, tools grid, install page, footer
- Responsive design for mobile
- Dark mode support via prefers-color-scheme

Copy the exact code from IMPLEMENT_PLAN.md section 1.6 (approximately lines 1583-1731).
```

#### Agent 4.3: PWA Essentials (manifest + sw.js + index.html)
```
Create THREE files:

1. `/home/richard/Documents/agentpmt/local_mcp/pwa/public/manifest.webmanifest`
- PWA manifest file
- Name: "Agent Payment"
- Icons referencing /agent-payment-logo.png
- Theme colors
- Standalone display mode

2. `/home/richard/Documents/agentpmt/local_mcp/pwa/public/sw.js`
- Service worker for offline support
- Cache management
- Fetch event handling with cache-first strategy

3. `/home/richard/Documents/agentpmt/local_mcp/pwa/index.html`
- HTML entry point
- Manifest link
- Service worker registration script
- Root div for React
- Meta tags

Copy the exact code from IMPLEMENT_PLAN.md section 1.2 (approximately lines 228-350).
```

---

### PHASE 5: Installation Scripts

#### Agent 5.1: install-macos.sh
```
Create the file `/home/richard/Documents/agentpmt/local_mcp/distribution/templates/install-macos.sh` exactly as specified in section 3.2 of IMPLEMENT_PLAN.md.

File content:
- Bash script for macOS installation
- Parameter: editor (claude/cursor/vscode)
- Config path determination per editor
- Copy executable to ~/.agent-payment
- Copy config.json
- JSON merging using jq (with fallback for manual config)
- Backup existing config
- Colored output (using ANSI codes)
- Error handling

Requirements:
- Must be executable (chmod +x)
- Must handle missing jq gracefully
- Must support both Claude Desktop and Cursor/VS Code config formats

Copy the exact code from IMPLEMENT_PLAN.md section 3.2 (install-macos.sh, approximately lines 2585-2744).
```

#### Agent 5.2: install-windows.ps1
```
Create the file `/home/richard/Documents/agentpmt/local_mcp/distribution/templates/install-windows.ps1` exactly as specified in section 3.2 of IMPLEMENT_PLAN.md.

File content:
- PowerShell script for Windows installation
- Parameter: -Editor (claude/cursor/vscode)
- Config path determination per editor
- Copy executable to ~/.agent-payment
- Copy config.json
- JSON merging using PowerShell ConvertFrom-Json/ConvertTo-Json
- Backup existing config
- Colored output (using Write-Host with colors)
- Error handling with try/catch

Requirements:
- Must handle execution policy restrictions
- Must use proper PowerShell cmdlets
- Must support both Claude Desktop and Cursor/VS Code config formats
- Must handle JSON depth correctly (-Depth 10)

Copy the exact code from IMPLEMENT_PLAN.md section 3.2 (install-windows.ps1, approximately lines 2746-2942).
```

#### Agent 5.3: install-linux.sh
```
Create the file `/home/richard/Documents/agentpmt/local_mcp/distribution/templates/install-linux.sh` exactly as specified in section 3.2 of IMPLEMENT_PLAN.md.

File content:
- Bash script for Linux installation
- Parameter: editor (claude/cursor/vscode)
- Config path determination per editor (uses ~/.config instead of ~/Library)
- Copy executable to ~/.agent-payment
- Copy config.json
- JSON merging using jq (with fallback for manual config)
- Backup existing config
- Colored output (using ANSI codes)
- Error handling

Requirements:
- Must be executable (chmod +x)
- Must handle missing jq gracefully
- Must use Linux-specific paths
- Must support both Claude Desktop and Cursor/VS Code config formats

Copy the exact code from IMPLEMENT_PLAN.md section 3.2 (install-linux.sh, approximately lines 2944-3103).
```

---

### PHASE 6: Build Scripts

#### Agent 6.1: build-all.sh
```
Create the file `/home/richard/Documents/agentpmt/local_mcp/scripts/build-all.sh` exactly as specified in section 3.1 of IMPLEMENT_PLAN.md.

File content:
- Bash script to build Go binaries for all platforms
- Cross-compilation commands for:
  - Windows (amd64)
  - macOS Intel (darwin/amd64)
  - macOS Apple Silicon (darwin/arm64)
  - Linux (amd64)
- Build flags: -ldflags="-s -w" for size optimization
- Output directory: ../distribution/binaries/
- Clean previous builds
- Echo build progress
- List built files with sizes

Requirements:
- Must be executable (chmod +x)
- Must work from mcp-server directory
- Must create output directories if they don't exist

Copy the exact code from IMPLEMENT_PLAN.md section 3.1 (approximately lines 2527-2583).
```

#### Agent 6.2: package-mcpb.sh
```
Create the file `/home/richard/Documents/agentpmt/local_mcp/scripts/package-mcpb.sh`.

File content:
- Bash script to create .mcpb packages for Claude Desktop
- For each platform: create ZIP with:
  - manifest.json (from templates/mcpb-manifest.json)
  - Binary (from binaries/<platform>/)
  - agent-payment-logo.png
  - README.md
- Rename ZIP to .mcpb extension
- Output to distribution/packages/

Requirements:
- Must be executable (chmod +x)
- Must use zip command
- Must handle multiple platforms
- Must include all required .mcpb contents

Create this script based on the .mcpb format described in IMPLEMENT_PLAN.md section 3.3 and the packaging needs.
```

#### Agent 6.3: package-installers.sh
```
Create the file `/home/richard/Documents/agentpmt/local_mcp/scripts/package-installers.sh`.

File content:
- Bash script to create installer ZIP packages
- For each platform: create ZIP with:
  - Binary (from binaries/<platform>/)
  - Install script (from templates/install-<platform>.sh or .ps1)
  - config.example.json
  - README.md
- Output to distribution/packages/
- Naming: agent-payment-<platform>-installer.zip

Requirements:
- Must be executable (chmod +x)
- Must handle all platforms (windows, macos-intel, macos-arm, linux)
- Must include correct install script per platform
- Must preserve executable permissions in ZIP

Create this script based on the installer packaging needs described in IMPLEMENT_PLAN.md.
```

---

### PHASE 7: CI/CD & Documentation

#### Agent 7.1: GitHub Actions Workflow
```
Create the file `/home/richard/Documents/agentpmt/local_mcp/.github/workflows/release.yml` exactly as specified in section 4.1 of IMPLEMENT_PLAN.md.

File content:
- GitHub Actions workflow for automated builds and releases
- Jobs:
  1. build-go: Build Go binaries for all platforms
  2. build-pwa: Build PWA and include binaries
  3. release: Create GitHub release with artifacts
  4. deploy-pwa: Deploy PWA to hosting
- Triggers: push to tags (v*), manual workflow_dispatch
- Uses: actions/checkout@v4, actions/setup-go@v5, actions/setup-node@v4, actions/upload-artifact@v4, actions/download-artifact@v4, softprops/action-gh-release@v1
- Cross-compilation in build-go job
- Artifact management between jobs

Copy the exact code from IMPLEMENT_PLAN.md section 4.1 (approximately lines 3105-3232).
```

#### Agent 7.2: Documentation Files
```
Create FOUR documentation files:

1. `/home/richard/Documents/agentpmt/local_mcp/README.md`
- Project overview
- Quick start guide
- Installation instructions
- Build instructions
- Links to other docs

2. `/home/richard/Documents/agentpmt/local_mcp/pwa/README.md`
- PWA-specific documentation
- Development setup
- Build and deployment
- Environment variables

3. `/home/richard/Documents/agentpmt/local_mcp/CONTRIBUTING.md`
- How to contribute
- Code style guidelines
- PR process
- Testing requirements

4. `/home/richard/Documents/agentpmt/local_mcp/LICENSE`
- MIT License (or your preferred license)

Create comprehensive documentation based on the project structure and implementation details from IMPLEMENT_PLAN.md.
```

---

## Execution Instructions

### Using This Guide

1. **Sequential tasks first**: Complete Phase 1 manually
   - Create directory structure
   - Run `npm create vite@latest pwa -- --template react-ts`
   - Run `cd mcp-server && go mod init github.com/your-org/agent-payment-server`

2. **Parallel tasks**: For each phase, launch ALL agents simultaneously in a single message:

```
I need you to launch the following agents in parallel:

PHASE 2: PWA Library Files
- Agent 2.1: [paste prompt for crypto.ts]
- Agent 2.2: [paste prompt for store.ts]
- Agent 2.3: [paste prompt for api.ts]

Please run all 3 agents at once and let me know when they're complete.
```

3. **Wait for completion**: Don't start next phase until current phase is complete

4. **Verify outputs**: After each phase, verify files were created correctly

---

## Efficiency Gains

### Without Parallel Execution
- **Estimated time**: 6-8 hours of sequential file creation
- **Bottleneck**: One task at a time

### With Parallel Execution
- **Estimated time**: 2-3 hours with proper parallelization
- **Speedup**: ~3x faster
- **Phases**:
  - Phase 1: 5 minutes (setup)
  - Phase 2: 15 minutes (3 agents in parallel)
  - Phase 3: 30 minutes (4 agents in parallel)
  - Phase 4: 30 minutes (3 agents in parallel)
  - Phase 5: 30 minutes (3 agents in parallel)
  - Phase 6: 20 minutes (3 agents in parallel)
  - Phase 7: 30 minutes (2 agents in parallel)
  - **Total**: ~2.5 hours

---

## Error Handling

If an agent fails:
1. Note which file(s) failed
2. Re-run just that specific agent
3. Don't wait for other successful agents

If multiple agents fail:
1. Check if it's a systemic issue (e.g., wrong base path)
2. Fix the issue
3. Re-run all failed agents in parallel

---

## Verification Checklist

After each phase:
- [ ] All files created in correct locations
- [ ] No syntax errors
- [ ] Imports are correct
- [ ] TypeScript compiles (for PWA files)
- [ ] Go compiles (for Go files)
- [ ] Scripts are executable

---

## Final Integration

After all phases complete:
1. Run `cd pwa && npm install`
2. Run `cd pwa && npm run build` to test PWA build
3. Run `cd mcp-server && go mod tidy && go build ./cmd/agent-payment-server` to test Go build
4. Run installation scripts to verify they work
5. Test complete workflow end-to-end

---

## Notes

- The Go MCP server code is already created in `/home/richard/Documents/agentpmt/local_mcp/mcp-server/` by the previous subagent research
- Focus parallel execution on the PWA and distribution files
- Package.json and tsconfig.json should be created after Phase 4 with correct dependencies
- agent-payment-logo.png should be placed in pwa/public/ before starting Phase 4.3
