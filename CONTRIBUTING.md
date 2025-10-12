# Contributing to Agent Payment MCP

Thank you for your interest in contributing! This document provides guidelines for contributing to the project.

## Code of Conduct

- Be respectful and inclusive
- Welcome newcomers and help them learn
- Focus on constructive feedback
- Assume positive intent

## Getting Started

### Prerequisites

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
git remote add upstream https://github.com/Apoth3osis-ai/agent-payment-mcp
```

### Setup Development Environment

```bash
# Install MCP Server dependencies
cd mcp-server
go mod download
cd ..

# Install Installer dependencies
cd installer
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
# Test MCP Server
cd mcp-server
go test ./...
go build ./cmd/agent-payment-server  # Ensure it builds
cd ..

# Test Installer
cd installer
go test ./...
go build ./cmd/installer  # Ensure it builds
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

### Go

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
func (c *Client) FetchTools(page, pageSize int) (*ToolsResponse, error) {
    // Implementation
}
```

### HTML/CSS/JavaScript (Installer Web UI)

- Keep HTML semantic and accessible
- Use vanilla JavaScript (no frameworks)
- Keep JavaScript simple and maintainable
- Use CSS variables for theming
- Support both light and dark modes
- Ensure mobile responsiveness

Example:

```css
.tool-card {
  background: var(--color-bg-secondary);
  padding: 1rem;
  border-radius: 8px;
}
```

## Testing Requirements

### Go Tests

- Write unit tests for all packages
- Test error handling
- Test API client with mock server
- Achieve >80% code coverage

Example:

```go
func TestFetchTools(t *testing.T) {
    // Test implementation
}
```

### Integration Tests

- Test installer flow end-to-end
- Test MCP server with real API (use test credentials)
- Test desktop client integration (manual)

## Documentation

- Update README.md if adding features
- Add GoDoc comments for new functions
- Update technical docs in `docs/development/` if changing architecture
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
Adds filtering capability to the MCP server tool list.

## Why
Improves performance when dealing with large tool sets.

## How
- Added caching layer to API client
- Implemented filtering in server registration
- Updated tool handler to use cache

## Testing
- Tested with 100+ tools
- Verified cache invalidation works correctly
- Checked memory usage
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
- Environment (OS, Go version, etc.)
- Error logs if applicable

### Feature Requests

Include:
- Clear description of feature
- Use case / motivation
- Proposed solution (if any)
- Alternatives considered

## Project Structure

```
agent-payment-mcp/
├── mcp-server/        # MCP server (Go)
│   ├── cmd/           # Main applications
│   ├── internal/      # Internal packages
│   └── go.mod
├── installer/         # Installer (Go)
│   ├── cmd/           # Installer binary
│   ├── internal/      # Installer logic & embedded web UI
│   └── go.mod
├── docs/              # Documentation
│   ├── INSTALLATION.md
│   └── development/   # Technical docs
└── README.md          # Customer-facing documentation
```

## Questions?

- Open a GitHub Issue
- Email: support@agentpmt.com
- Website: [agentpmt.com](https://agentpmt.com)

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

Thank you for contributing! 🎉
