We are building a **PWA UI** that’s also an **MCP client (FastMCP)**. Users install the app, paste their **API key** + **Budget key**, then browse/execute tools that are proxied by your hosted endpoints:

* **GET** `https://api.agentpmt.com/products/fetch` → list tools
* **POST** `https://api.agentpmt.com/products/purchase` → run a tool

Below is a build plan with file stubs. It keeps secrets local (encrypted in IndexedDB), supports offline install, and maps the remote tools into MCP “function tools.”

---

# Project brief

* Framework: **Vite + React + TypeScript** (small, fast)
* MCP client: **FastMCP** (runs in the browser)
* Installable: **PWA** (manifest + service worker)
* Secrets: stored **encrypted in IndexedDB** using **WebCrypto (AES-GCM)** with a locally cached random key (optionally protect with a passcode)
* Networking: plain `fetch` with the two required headers
* Logo: save your PNG at **`public/agent-payment-logo.png`** (the path the PWA will use)

---

# Commands

```bash
npm create vite@latest agent-payment-pwa -- --template react-ts
cd agent-payment-pwa
npm i fastmcp idb
# FastMCP type helpers (if published separately, otherwise remove this line)
# npm i @types/serviceworker
```

---

# File layout (only the files we add/modify)

```
agent-payment-pwa/
  public/
    agent-payment-logo.png        # <-- put your provided logo here
    manifest.webmanifest
    sw.js
  src/
    main.tsx
    App.tsx
    routes/Settings.tsx
    routes/Tools.tsx
    components/Header.tsx
    lib/crypto.ts
    lib/store.ts
    lib/api.ts
    mcp/fastmcpClient.ts
    styles.css
  index.html
  vite.config.ts
```

---

# PWA essentials

**public/manifest.webmanifest**

```json
{
  "name": "Agent Payment",
  "short_name": "AgentPay",
  "display": "standalone",
  "start_url": "/",
  "background_color": "#ffffff",
  "theme_color": "#0b0b0c",
  "icons": [
    { "src": "/agent-payment-logo.png", "sizes": "192x192", "type": "image/png" },
    { "src": "/agent-payment-logo.png", "sizes": "512x512", "type": "image/png" }
  ]
}
```

**public/sw.js** (minimal offline shell)

```js
self.addEventListener("install", (e) => {
  e.waitUntil(caches.open("ap-shell-v1").then(c => c.addAll(["/", "/index.html"])));
});
self.addEventListener("fetch", (e) => {
  e.respondWith(
    caches.match(e.request).then(r => r || fetch(e.request))
  );
});
```

**index.html** (hook up manifest + SW)

```html
<!doctype html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <link rel="manifest" href="/manifest.webmanifest" />
  <meta name="viewport" content="width=device-width,initial-scale=1" />
  <title>Agent Payment</title>
</head>
<body>
  <div id="root"></div>
  <script type="module" src="/src/main.tsx"></script>
  <script>
    if ("serviceWorker" in navigator) {
      window.addEventListener("load", () => {
        navigator.serviceWorker.register("/sw.js");
      });
    }
  </script>
</body>
</html>
```

---

# Secrets: local, encrypted storage

**src/lib/crypto.ts**

```ts
// AES-GCM helpers (WebCrypto)
export async function genKey(): Promise<CryptoKey> {
  return crypto.subtle.generateKey({ name: "AES-GCM", length: 256 }, true, ["encrypt","decrypt"]);
}
export async function importKey(raw: ArrayBuffer): Promise<CryptoKey> {
  return crypto.subtle.importKey("raw", raw, "AES-GCM", true, ["encrypt","decrypt"]);
}
export async function exportKey(key: CryptoKey): Promise<ArrayBuffer> {
  return crypto.subtle.exportKey("raw", key);
}
export async function encryptJSON(key: CryptoKey, data: unknown): Promise<{iv: Uint8Array, buf: ArrayBuffer}> {
  const iv = crypto.getRandomValues(new Uint8Array(12));
  const enc = new TextEncoder().encode(JSON.stringify(data));
  const buf = await crypto.subtle.encrypt({ name: "AES-GCM", iv }, key, enc);
  return { iv, buf };
}
export async function decryptJSON<T>(key: CryptoKey, iv: Uint8Array, buf: ArrayBuffer): Promise<T> {
  const dec = await crypto.subtle.decrypt({ name: "AES-GCM", iv }, key, buf);
  return JSON.parse(new TextDecoder().decode(new Uint8Array(dec)));
}
```

**src/lib/store.ts**

```ts
import { openDB } from "idb";
import { genKey, importKey, exportKey, encryptJSON, decryptJSON } from "./crypto";

const DB = "agentpay-db";
const STORE = "secrets";
const KEY_SLOT = "k";

type SecretBundle = { apiKey: string; budgetKey: string; auth?: string };

async function db() {
  return openDB(DB, 1, {
    upgrade(db) {
      db.createObjectStore(STORE);
    }
  });
}

// persist a generated symmetric key in localStorage (device-scoped)
async function getOrCreateSymKey(): Promise<CryptoKey> {
  const hex = localStorage.getItem("agentpay_symkey");
  if (hex) {
    const raw = new Uint8Array(hex.match(/.{1,2}/g)!.map(h => parseInt(h,16))).buffer;
    return importKey(raw);
  }
  const key = await genKey();
  const raw = new Uint8Array(await exportKey(key));
  const hexOut = Array.from(raw).map(b => b.toString(16).padStart(2,"0")).join("");
  localStorage.setItem("agentpay_symkey", hexOut);
  return key;
}

export async function saveSecrets(bundle: SecretBundle) {
  const key = await getOrCreateSymKey();
  const {iv, buf} = await encryptJSON(key, bundle);
  const d = await db();
  await d.put(STORE, { iv: Array.from(iv), buf: Array.from(new Uint8Array(buf)) }, KEY_SLOT);
}

export async function loadSecrets(): Promise<SecretBundle | null> {
  const d = await db();
  const rec = await d.get(STORE, KEY_SLOT);
  if (!rec) return null;
  const key = await getOrCreateSymKey();
  const iv = new Uint8Array(rec.iv);
  const buf = new Uint8Array(rec.buf).buffer;
  return decryptJSON<SecretBundle>(key, iv, buf);
}

export async function clearSecrets() {
  const d = await db();
  await d.delete(STORE, KEY_SLOT);
}
```

> ⚠️ Note: client-side apps can’t be perfectly “secret.” This scheme encrypts at rest and avoids plaintext storage; it’s appropriate for a PWA UX. For shared devices, add a passcode screen and derive the key with PBKDF2.

---

# API glue to your endpoints

**src/lib/api.ts**

```ts
const BASE = "https://api.agentpmt.com"; // swap to testnet.* when needed

export type ToolRecord = {
  type: "function";
  function: {
    name: string;
    description: string;
    parameters: Record<string, unknown>;
  };
  ["x-prepaid-balance"]?: unknown;
  ["x-pricing"]?: unknown;
};

export type FetchResponse = {
  success: boolean;
  preprompt?: string;
  example?: unknown;
  tools: ToolRecord[];
  pagination?: unknown;
};

export async function fetchTools(keys: {apiKey: string; budgetKey: string; auth?: string}, page=1, pageSize=20): Promise<FetchResponse> {
  const url = `${BASE}/products/fetch?page=${page}&page_size=${pageSize}`;
  const r = await fetch(url, {
    method: "GET",
    headers: {
      "Content-Type": "application/json",
      "x-api-key": keys.apiKey,
      "x-budget-key": keys.budgetKey,
      ...(keys.auth ? { "Authorization": keys.auth } : {})
    }
  });
  if (!r.ok) throw new Error(`Fetch tools failed: ${r.status}`);
  return r.json();
}

export async function purchaseTool(
  keys: {apiKey: string; budgetKey: string},
  product_id: string,
  parameters: Record<string, unknown>
) {
  const r = await fetch(`${BASE}/products/purchase`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "x-api-key": keys.apiKey,
      "x-budget-key": keys.budgetKey
    },
    body: JSON.stringify({ product_id, parameters })
  });
  if (!r.ok) throw new Error(`Purchase failed: ${r.status}`);
  return r.json();
}
```

---

# FastMCP bridge (turn remote tools into MCP tools)

**src/mcp/fastmcpClient.ts**

```ts
import { fetchTools, purchaseTool } from "../lib/api";
import { loadSecrets } from "../lib/store";

// Pseudo-API for FastMCP (keep generic if exact names differ)
type McpTool = {
  name: string;
  description?: string;
  inputSchema: object;
  handler: (args: Record<string, unknown>) => Promise<unknown>;
};

export async function buildMcpTools(): Promise<McpTool[]> {
  const secrets = await loadSecrets();
  if (!secrets) return [];
  const { tools } = await fetchTools(secrets);
  return tools.map(t => {
    const fn = t.function;
    return {
      name: fn.name,
      description: fn.description,
      inputSchema: fn.parameters,
      handler: async (args) => {
        // Map MCP tool call -> POST /products/purchase
        const res = await purchaseTool({ apiKey: secrets.apiKey, budgetKey: secrets.budgetKey }, fn.name, args ?? {});
        return res;
      }
    };
  });
}

// Example initializer: expose an MCP client that lists these tools
export async function initFastMcpClient(fastmcp: any) {
  const toolDefs = await buildMcpTools();
  // Register tools with FastMCP
  for (const t of toolDefs) {
    fastmcp.registerTool({
      name: t.name,
      description: t.description,
      parameters: t.inputSchema,
      handler: t.handler
    });
  }
}
```

> If the FastMCP API you use has different method names, keep the mapping exactly the same: **name + parameters → handler(args) → purchaseTool**. The important part is that the **tool schema comes from `GET /products/fetch`** and execution goes to **`POST /products/purchase`**.

---

# UI (settings + tools)

**src/components/Header.tsx**

```tsx
export default function Header() {
  return (
    <header className="app-header">
      <img src="/agent-payment-logo.png" alt="Agent Payment" style={{height: 28}} />
      <span style={{fontWeight: 700, marginLeft: 8}}>agent PAYMENT</span>
    </header>
  );
}
```

**src/routes/Settings.tsx**

```tsx
import { useEffect, useState } from "react";
import { saveSecrets, loadSecrets, clearSecrets } from "../lib/store";

export default function Settings() {
  const [apiKey, setApiKey] = useState("");
  const [budgetKey, setBudgetKey] = useState("");
  const [auth, setAuth] = useState(""); // optional

  useEffect(() => {
    loadSecrets().then(s => { if (s) { setApiKey(s.apiKey); setBudgetKey(s.budgetKey); setAuth(s.auth ?? ""); }});
  }, []);

  return (
    <div className="card">
      <h2>Connect to Agent Payment</h2>
      <p>Enter your API and Budget keys. They are stored locally in this PWA, encrypted at rest.</p>
      <label>API Key<input value={apiKey} onChange={e=>setApiKey(e.target.value)} placeholder="x-api-key" /></label>
      <label>Budget Key<input value={budgetKey} onChange={e=>setBudgetKey(e.target.value)} placeholder="x-budget-key" /></label>
      <label>Authorization (optional)<input value={auth} onChange={e=>setAuth(e.target.value)} placeholder="Bearer ..." /></label>
      <div className="row">
        <button onClick={() => saveSecrets({ apiKey, budgetKey, auth: auth || undefined })}>Save</button>
        <button onClick={() => clearSecrets()}>Clear</button>
      </div>
    </div>
  );
}
```

**src/routes/Tools.tsx**

```tsx
import { useEffect, useState } from "react";
import { fetchTools, purchaseTool } from "../lib/api";
import { loadSecrets } from "../lib/store";

export default function Tools() {
  const [tools, setTools] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [result, setResult] = useState<any>(null);

  useEffect(() => {
    (async () => {
      const secrets = await loadSecrets();
      if (!secrets) { setLoading(false); return; }
      try {
        const data = await fetchTools(secrets, 1, 20);
        setTools(data.tools);
      } finally { setLoading(false); }
    })();
  }, []);

  if (loading) return <p>Loading tools…</p>;
  if (!tools.length) return <p>No tools yet. Add keys in Settings.</p>;

  return (
    <div className="tools-grid">
      {tools.map(t => (
        <ToolCard key={t.function.name} t={t} onRun={setResult} />
      ))}
      {result && (
        <pre className="result">{JSON.stringify(result, null, 2)}</pre>
      )}
    </div>
  );
}

function ToolCard({ t, onRun }: { t:any; onRun: (r:any)=>void }) {
  const schema = t.function.parameters?.properties ?? {};
  const req: string[] = t.function.parameters?.required ?? [];
  const [values, setValues] = useState<Record<string, any>>({});

  function update(k:string, v:string) { setValues(s => ({...s, [k]: v})); }

  async function run() {
    const secrets = await loadSecrets();
    if (!secrets) return alert("Add keys in Settings");
    const res = await purchaseTool({ apiKey: secrets.apiKey, budgetKey: secrets.budgetKey }, t.function.name, values);
    onRun(res);
  }

  return (
    <div className="card">
      <h3>{t.function.name}</h3>
      <p>{t.function.description}</p>
      {Object.entries(schema).map(([k, def]: any) => (
        <label key={k}>
          {k}{req.includes(k) ? " *" : ""} — <small>{def?.description || ""}</small>
          <input onChange={e => update(k, e.target.value)} placeholder={k} />
        </label>
      ))}
      <button onClick={run}>Run Tool</button>
    </div>
  );
}
```

**src/App.tsx**

```tsx
import { useEffect } from "react";
import Header from "./components/Header";
import Settings from "./routes/Settings";
import Tools from "./routes/Tools";
// Optional: wire FastMCP once your UI loads
import { initFastMcpClient } from "./mcp/fastmcpClient";

declare global { interface Window { fastmcp?: any } }

export default function App() {
  useEffect(() => {
    if (window.fastmcp) initFastMcpClient(window.fastmcp); // no-op if not present
  }, []);
  return (
    <div className="app">
      <Header />
      <main>
        <Settings />
        <hr />
        <Tools />
      </main>
    </div>
  );
}
```

**src/main.tsx**

```tsx
import React from "react";
import ReactDOM from "react-dom/client";
import App from "./App";
import "./styles.css";

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode><App /></React.StrictMode>
);
```

**src/styles.css** (tiny, pleasant defaults)

```css
:root { color-scheme: light dark; font: 16px system-ui, sans-serif; }
.app { max-width: 980px; margin: 0 auto; padding: 1.25rem; }
.app-header { display:flex; align-items:center; gap:.5rem; padding:.5rem 0; }
.card { border: 1px solid color-mix(in oklab, CanvasText 10%, transparent);
        border-radius: 16px; padding: 1rem; margin: .75rem 0; }
label { display:block; margin:.5rem 0; }
input { width:100%; padding:.5rem; border-radius: 10px; }
button { padding:.5rem .8rem; border-radius: 12px; cursor:pointer; }
.tools-grid { display:grid; grid-template-columns: repeat(auto-fill,minmax(280px,1fr)); gap: .75rem; }
.result { white-space: pre-wrap; background: color-mix(in oklab, Canvas 95%, CanvasText 5%);
          padding: .75rem; border-radius: 12px; }
```

---

# Wiring FastMCP in the browser

You have two options:

1. **Bundle FastMCP** and expose a global (as shown with `window.fastmcp`) then call `initFastMcpClient`.
2. If FastMCP ships as an ESM client, import and pass it in directly:

```ts
// Example pattern
import { FastMCP } from "fastmcp";
useEffect(() => {
  const client = new FastMCP();
  initFastMcpClient(client);
}, []);
```

> The only requirement: **register one MCP tool per item returned by `GET /products/fetch`** and have each tool call **`purchaseTool(..)`** with `{ product_id: tool.function.name, parameters }`.

---

# CORS & environments

* Ensure `https://api.agentpmt.com` (and `https://testnet.api.agentpmt.com` if used) allows your PWA origin via CORS.
* In development, Vite runs on `http://localhost:5173` → allow this origin too.
* You can flip between prod/testnet by switching `BASE` in `api.ts` (or use an environment toggle).

---

# Logo placement

* Put your provided PNG at **`public/agent-payment-logo.png`** (root project’s `public/`).
* It’s referenced by the header, manifest icons, and the browser will pick it up for the installable PWA.

---

# What the agent must deliver

* [ ] PWA compiles & installs (manifest + SW)
* [ ] Settings screen persists **x-api-key** and **x-budget-key** locally (encrypted)
* [ ] Tools screen fetches tool list and renders inputs from the JSON schema
* [ ] Clicking **Run Tool** posts to `/products/purchase` and shows the JSON result
* [ ] FastMCP client is initialized and exposes the same tools to the MCP layer
* [ ] Clear readme with `npm run dev` and `npm run build`

---
Short answer: yes—we can make it “one-click-ish.” Your PWA can generate the exact config files/scripts that desktop MCP clients look for, and for Claude Desktop specifically there’s now a true one-click path (Desktop Extensions). Here’s the plan that a coding LLM can implement right now.

# What each client looks for (so we can auto-provision)

* **Claude Desktop** reads a JSON file you can edit or create at:
  macOS: `~/Library/Application Support/Claude/claude_desktop_config.json`
  Windows: `%APPDATA%\Claude\claude_desktop_config.json` ([Model Context Protocol][1])
* **Cursor** supports a **global** config at `~/.cursor/mcp.json` and a **project** config at `<project>/.cursor/mcp.json`. ([Cursor][2])
* **VS Code (GitHub Copilot Chat)** supports **workspace** config at `<project>/.vscode/mcp.json` and **user/global** via the “MCP: Add Server” command (opens a user-profile `mcp.json`). ([Visual Studio Code][3])

# “Install to your editor” buttons in the PWA

Add an **“Install to your editor”** section with three big buttons: “Claude,” “Cursor,” “VS Code.” Clicking each will:

1. **Claude Desktop – two options**

   * **One-click (recommended): Desktop Extension (.mcpb).** Package your server as a **Claude Desktop Extension** so users click *Install* in Claude and it wires everything automatically. Your PWA can host the `.mcpb` file and link to `Settings → Extensions → Install Extension…` in Claude. ([FastMCP][4])

   * **Classic config:** Offer a per-OS script that writes/merges `claude_desktop_config.json` with an entry for your local FastMCP proxy (stdio). ([Model Context Protocol][1])

     * **macOS (bash) script template** your PWA downloads:

       ```bash
       #!/usr/bin/env bash
       set -euo pipefail
       CFG="${HOME}/Library/Application Support/Claude/claude_desktop_config.json"
       mkdir -p "$(dirname "$CFG")"
       # Write/merge JSON to add your server:
       # {
       #   "mcpServers": {
       #     "agent-payment": {
       #       "command": "uvx",
       #       "args": ["fastmcp", "run", "--stdio", "/PATH/TO/agent_payment_server.py"],
       #       "env": {"AGENTPAY_API_KEY":"<redacted>", "AGENTPAY_BUDGET_KEY":"<redacted>"}
       #     }
       #   }
       # }
       ```

       (Windows PowerShell variant does the same to `%APPDATA%\Claude\claude_desktop_config.json`.) Paths and env come from the user’s saved keys in your PWA.

   * **Bonus:** If the user has FastMCP ≥ **2.10.3**, you can show a copy-paste one-liner they can run that auto-installs into Claude, e.g.:
     `fastmcp install claude-desktop path/to/agent_payment_server.py` ([gofastmcp.com][5])

2. **Cursor**

   * Offer a downloadable **`mcp.json`** with the server definition (stdio or http), and a short script that puts it at `~/.cursor/mcp.json` (or instructs to drop into `.cursor/mcp.json` in the project). ([Cursor][2])

3. **VS Code**

   * Offer either a **workspace** file download to place at `.vscode/mcp.json` **or** a snippet for the **user/global** config (the PWA can’t write it directly, but can show a “Copy JSON” button and short steps to run “MCP: Add Server” to paste it). ([Visual Studio Code][3])

> Note on permissions: a PWA can’t write to those folders itself. That’s why we generate **ready-to-save files** and **tiny per-OS scripts** (bash/PowerShell) the user runs once. On Claude, the **.mcpb extension** path removes even that friction. ([FastMCP][4])

# The server entry you generate (stdio)

Use a tiny local **FastMCP “proxy” server** that maps each tool from your API into MCP tools (exactly like we designed). Then register it as **stdio** so every client discovers it the same way.

**Claude Desktop config shape** (merge into `claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "agent-payment": {
      "command": "uvx",
      "args": ["fastmcp", "run", "--stdio", "/ABS/PATH/agent_payment_server.py"],
      "env": {
        "AGENTPAY_API_KEY": "••••",
        "AGENTPAY_BUDGET_KEY": "••••"
      }
    }
  }
}
```

Claude supports `command` + `args` (stdio) just like this. ([Model Context Protocol][1])

**Cursor global `~/.cursor/mcp.json`** (or project `.cursor/mcp.json`):

```json
{
  "servers": {
    "agent-payment": {
      "type": "stdio",
      "command": "uvx",
      "args": ["fastmcp", "run", "--stdio", "/ABS/PATH/agent_payment_server.py"],
      "env": {
        "AGENTPAY_API_KEY": "••••",
        "AGENTPAY_BUDGET_KEY": "••••"
      }
    }
  }
}
```

([Cursor][2])

**VS Code workspace `.vscode/mcp.json`**:

```json
{
  "servers": {
    "agent-payment": {
      "type": "stdio",
      "command": "uvx",
      "args": ["fastmcp", "run", "--stdio", "/ABS/PATH/agent_payment_server.py"],
      "env": {
        "AGENTPAY_API_KEY": "••••",
        "AGENTPAY_BUDGET_KEY": "••••"
      }
    }
  }
}
```

(Alternatively, users can add this via **Command Palette → “MCP: Add Server”** for global config.) ([Visual Studio Code][3])

> If you ship a pure **remote HTTP** server instead: VS Code/others accept `"type": "http", "url": "https://…"` in `mcp.json`. But for Cursor and Claude Desktop, **stdio** is the most broadly compatible local-install experience. ([Visual Studio Code][3])

# How the PWA makes this easy (exact UX)

* **Settings**: user pastes API key + Budget key (we already store them safely in the PWA).
* **Install panel**:

  * **Claude**: two tiles: “Install via Extension (.mcpb)” and “Install via Config.”

    * **Extension tile**: downloads your `.mcpb` and shows the 3-step Claude UI (“Settings → Extensions → Install Extension…”). ([Claude Help Center][6])
    * **Config tile**: lets the user pick OS → downloads the right script + JSON (pre-filled with their keys) and shows “Run this once.”
  * **Cursor**: dropdown “Global vs Project” → download `mcp.json` (pre-filled) + copy path to place file. ([Cursor][2])
  * **VS Code**: toggle “Workspace” vs “Global” → download `.vscode/mcp.json` or show copy-paste for Command Palette flow. ([Visual Studio Code][3])
* **Advanced**: show a copy-paste **FastMCP CLI** command:
  `fastmcp install claude-desktop /ABS/PATH/agent_payment_server.py` (auto-wires Claude). ([gofastmcp.com][5])

# Notes on FastMCP Cloud “buttons”

FastMCP has **Cloud deployment** guides and a **CLI** that can *install servers into Claude Desktop automatically* (great to emulate as a “button” by just showing the one-liner). If you later host a landing page, you can surface the same “Install to Claude Desktop” action via that CLI or via a downloadable **.mcpb** Desktop Extension. ([gofastmcp.com][5])

---

## TL;DR for the implementer

1. Keep our PWA as the friendly hub for keys + tool browsing.
2. Add an **Install** screen that:

   * Generates **pre-filled** config files/scripts for **Claude, Cursor, VS Code**.
   * Offers a **.mcpb** package for **one-click Claude install**, and shows the **FastMCP CLI** one-liner as an alternative. ([FastMCP][4])
3. Our local **FastMCP stdio proxy** (started by those configs) simply fetches tools from `GET /products/fetch` and calls `POST /products/purchase` when invoked—exactly the flow we already designed.

If you want, I can add the exact bash/PowerShell scripts and the minimal `.mcpb` manifest next.

[1]: https://modelcontextprotocol.io/docs/develop/connect-local-servers "Connect to local MCP servers - Model Context Protocol"
[2]: https://cursor.com/docs/context/mcp?utm_source=chatgpt.com "Model Context Protocol (MCP) | Cursor Docs"
[3]: https://code.visualstudio.com/docs/copilot/customization/mcp-servers?utm_source=chatgpt.com "Use MCP servers in VS Code"
[4]: https://fastmcp.me/MCP/Details/80/mcp-installer?utm_source=chatgpt.com "MCP Installer MCP - 1-Click Ready | FastMCP"
[5]: https://gofastmcp.com/integrations/claude-desktop?utm_source=chatgpt.com "Claude Desktop 🤝 FastMCP"
[6]: https://support.claude.com/en/articles/10949351-getting-started-with-local-mcp-servers-on-claude-desktop?utm_source=chatgpt.com "Getting Started with Local MCP Servers on Claude Desktop"
