# Agent Payment PWA

Progressive Web App for browsing Agent Payment tools and generating installers for desktop clients.

## Overview

The PWA provides a user-friendly interface to:
- Enter and store API credentials (encrypted locally)
- Browse available tools from the Agent Payment API
- Test tools directly in the browser
- Download installer packages for Claude Desktop, Cursor, and VS Code

## Technology Stack

- **Framework**: React 18 + TypeScript
- **Build Tool**: Vite
- **Storage**: IndexedDB (via `idb` library)
- **Encryption**: WebCrypto API (AES-GCM)
- **Packaging**: JSZip
- **Styling**: Custom CSS with CSS variables for theming

## Features

### Security
- API credentials encrypted at rest using AES-GCM
- Symmetric key stored in localStorage (browser-specific)
- No credentials sent to any server except Agent Payment API

### Offline Support
- Service Worker for offline functionality
- Cached assets for faster loading
- Progressive enhancement

### Responsive Design
- Mobile-friendly interface
- Dark/light mode support (system preference)
- Accessible components

## Development

### Prerequisites

- Node.js 20+
- npm or yarn

### Setup

```bash
# Install dependencies
npm install

# Run development server
npm run dev

# Visit http://localhost:5173
```

### Project Structure

```
pwa/
├── src/
│   ├── components/
│   │   └── Header.tsx          # App header with logo
│   ├── routes/
│   │   ├── Settings.tsx        # API credentials entry
│   │   ├── Tools.tsx           # Tool browser & tester
│   │   └── Install.tsx         # Installer generator
│   ├── lib/
│   │   ├── crypto.ts           # WebCrypto utilities
│   │   ├── store.ts            # IndexedDB storage
│   │   └── api.ts              # REST API client
│   ├── App.tsx                 # Main app component
│   ├── main.tsx                # Entry point
│   └── styles.css              # Global styles
├── public/
│   ├── agent-payment-logo.png  # Logo asset
│   ├── manifest.webmanifest    # PWA manifest
│   └── sw.js                   # Service worker
├── index.html                  # HTML entry point
├── vite.config.ts              # Vite configuration
├── package.json
└── tsconfig.json
```

## Building

```bash
# Production build
npm run build

# Output: dist/

# Preview production build
npm run preview
```

## Environment Variables

Create `.env` for custom configuration:

```env
# API base URL (optional, defaults to https://api.agentpmt.com)
VITE_API_BASE_URL=https://api.agentpmt.com
```

## Deployment

The PWA is a static site and can be deployed to any hosting service:

### Vercel

```bash
npm install -g vercel
vercel deploy
```

### Netlify

```bash
npm install -g netlify-cli
netlify deploy --prod --dir=dist
```

### AWS S3

```bash
aws s3 sync dist/ s3://your-bucket-name --delete
```

### GitHub Pages

Use the provided GitHub Actions workflow in `.github/workflows/release.yml`.

## Testing

### Unit Tests

```bash
npm run test
```

### E2E Tests

```bash
npm run test:e2e
```

### Manual Testing

1. **Settings Page**
   - Enter API credentials
   - Save and verify encryption
   - Clear credentials

2. **Tools Page**
   - Browse available tools
   - Expand tool details
   - Test tool execution with parameters

3. **Install Page**
   - Select editor (Claude/Cursor/VS Code)
   - Select platform (Windows/macOS/Linux)
   - Choose installation method (.mcpb or script)
   - Download package
   - Verify package contents

## Security Considerations

### Credential Storage

- Credentials are encrypted using AES-GCM (256-bit)
- Encryption key stored in localStorage (origin-specific)
- Data stored in IndexedDB (origin-specific)
- Credentials never sent to any server except Agent Payment API

### Content Security Policy

The PWA implements strict CSP:
- No inline scripts (except initial SW registration)
- No eval or unsafe-eval
- Only connect to Agent Payment API

### HTTPS Only

The PWA requires HTTPS in production (enforced by Service Worker).

## API Integration

### Endpoints Used

- `GET /products/fetch` - Fetch available tools
- `POST /products/purchase` - Execute a tool

### Authentication

Requests include:
- `x-api-key`: User's API key
- `x-budget-key`: User's budget key
- `Authorization`: Optional additional auth (if configured)

## Browser Support

- Chrome/Edge 90+
- Firefox 88+
- Safari 14+
- Opera 76+

Features gracefully degrade on older browsers.

## Troubleshooting

### Service Worker Issues

```bash
# Clear service worker cache
# In browser DevTools: Application > Storage > Clear site data
```

### Build Issues

```bash
# Clear cache and reinstall
rm -rf node_modules package-lock.json
npm install
```

### TypeScript Errors

```bash
# Check types
npm run type-check
```

## Contributing

See [CONTRIBUTING.md](../CONTRIBUTING.md) for contribution guidelines.

## License

MIT
