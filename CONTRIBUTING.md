# Contributing to Agent Payment MCP

Thank you for your interest in contributing! This document provides guidelines for contributing to the project.

## Code of Conduct

- Be respectful and inclusive
- Welcome newcomers and help them learn
- Focus on constructive feedback
- Assume positive intent

## Getting Started

### Prerequisites

- Node.js 20+
- Go 1.21+
- Git
- A GitHub account

### Fork and Clone

```bash
# Fork the repository on GitHub
# Clone your fork
git clone https://github.com/YOUR-USERNAME/agent-payment-mcp
cd agent-payment-mcp

# Add upstream remote
git remote add upstream https://github.com/your-org/agent-payment-mcp
```

### Setup Development Environment

```bash
# Install PWA dependencies
cd pwa
npm install
cd ..

# Install Go dependencies
cd mcp-server
go mod download
cd ..
```

## Development Workflow

### 1. Create a Branch

```bash
git checkout -b feature/your-feature-name
# or
git checkout -b fix/your-bug-fix
```

Branch naming conventions:
- `feature/*` - New features
- `fix/*` - Bug fixes
- `docs/*` - Documentation updates
- `refactor/*` - Code refactoring
- `test/*` - Test additions/fixes

### 2. Make Changes

Follow our code style guidelines (see below).

### 3. Test Your Changes

```bash
# Test PWA
cd pwa
npm run test
npm run build  # Ensure it builds
cd ..

# Test Go server
cd mcp-server
go test ./...
go build ./cmd/agent-payment-server  # Ensure it builds
cd ..
```

### 4. Commit Changes

Write clear, descriptive commit messages:

```bash
git add .
git commit -m "feat: add new tool filtering feature"
# or
git commit -m "fix: resolve issue with API credential storage"
```

Commit message format:
- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation changes
- `style:` - Code style changes (formatting, etc.)
- `refactor:` - Code refactoring
- `test:` - Test additions/changes
- `chore:` - Build process, dependencies, etc.

### 5. Push and Create Pull Request

```bash
git push origin feature/your-feature-name
```

Then create a Pull Request on GitHub.

## Code Style Guidelines

### TypeScript/React (PWA)

- Use TypeScript for type safety
- Follow React hooks best practices
- Use functional components
- Keep components small and focused
- Use meaningful variable and function names
- Add JSDoc comments for complex functions
- Run `npm run lint` before committing

Example:

```typescript
/**
 * Encrypts JSON data using AES-GCM
 * @param key - CryptoKey for encryption
 * @param data - Data to encrypt
 * @returns Encrypted data with IV
 */
export async function encryptJSON(
  key: CryptoKey,
  data: unknown
): Promise<{ iv: Uint8Array; ciphertext: ArrayBuffer }> {
  // Implementation
}
```

### Go (MCP Server)

- Follow standard Go conventions
- Use `gofmt` for formatting
- Add comments for exported functions
- Keep functions small and focused
- Use meaningful variable names
- Handle errors explicitly
- Run `go fmt ./...` and `go vet ./...` before committing

Example:

```go
// FetchTools retrieves available tools from the Agent Payment API
func (c *Client) FetchTools(ctx context.Context) ([]Tool, error) {
    // Implementation
}
```

### CSS

- Use CSS variables for theming
- Follow BEM naming when appropriate
- Keep selectors specific but not overly nested
- Support both light and dark modes
- Use semantic class names

Example:

```css
.tool-card {
  background: var(--color-bg-secondary);
}

.tool-card__title {
  font-size: 1.25rem;
}
```

## Testing Requirements

### PWA Tests

- Write unit tests for utility functions
- Test React components with React Testing Library
- Test API client with mock responses
- Ensure crypto functions work correctly

### Go Tests

- Write unit tests for all packages
- Test error handling
- Test API client with mock server
- Achieve >80% code coverage

### Integration Tests

- Test full installer generation flow
- Test Go server with real API (use test credentials)
- Test desktop client integration (manual)

## Documentation

- Update README.md if adding features
- Add JSDoc/GoDoc comments for new functions
- Update IMPLEMENT_PLAN.md if changing architecture
- Include examples in documentation

## Pull Request Guidelines

### Before Submitting

- [ ] Code follows style guidelines
- [ ] All tests pass
- [ ] No new warnings or errors
- [ ] Documentation is updated
- [ ] Commits are clean and descriptive

### PR Description

Include:
1. **What** - What does this PR do?
2. **Why** - Why is this change needed?
3. **How** - How does it work?
4. **Testing** - How was it tested?
5. **Screenshots** - For UI changes

Example:

```markdown
## What
Adds filtering capability to the Tools page to search tools by name.

## Why
Users requested ability to quickly find tools in a long list.

## How
- Added search input component
- Implemented client-side filtering
- Updated Tools component to filter results

## Testing
- Tested with 50+ tools
- Verified case-insensitive search
- Checked mobile responsiveness

## Screenshots
[Include screenshots here]
```

### Review Process

1. Maintainer reviews code
2. CI/CD runs automated tests
3. Feedback is provided
4. You make requested changes
5. PR is approved and merged

## Reporting Issues

### Bug Reports

Include:
- Clear description of the bug
- Steps to reproduce
- Expected vs actual behavior
- Environment (OS, browser, versions)
- Screenshots if applicable

### Feature Requests

Include:
- Clear description of feature
- Use case / motivation
- Proposed solution (if any)
- Alternatives considered

## Questions?

- Open a GitHub Discussion
- Join our community chat
- Email: support@agentpmt.com

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

Thank you for contributing! 🎉
