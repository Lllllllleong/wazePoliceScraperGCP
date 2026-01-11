# Contributing

Contributions are welcome. This document outlines the development workflow.

## Prerequisites

- Go 1.24+
- Node.js 20+
- Docker
- Google Cloud SDK (`gcloud`)
- Firebase CLI (`firebase-tools`)
- GitHub CLI (`gh`) - optional

## Getting Started

1. Fork the repository
2. Clone your fork:
   ```bash
   git clone https://github.com/YOUR_USERNAME/wazePoliceScraperGCP.git
   cd wazePoliceScraperGCP
   ```
3. Add upstream remote:
   ```bash
   git remote add upstream https://github.com/Lllllllleong/wazePoliceScraperGCP.git
   ```

## Development Workflow

### 1. Create a Branch

Use descriptive branch names with prefixes:

```bash
git checkout -b feat/your-feature-name
git checkout -b fix/bug-description
git checkout -b docs/documentation-update
```

### 2. Make Changes

Follow existing code patterns and style:

**Go Code:**
- Run linter: `golangci-lint run`
- Format code: `gofmt -s -w .`
- Run tests: `go test -v -race ./...`

**Frontend:**
- Lint: `cd dataAnalysis && npm run lint` (if configured)
- Format: Follow existing code style
- Test: `cd dataAnalysis && npm test`

### 3. Test Your Changes

All tests must pass before submitting:

**Backend:**
```bash
# Unit tests
go test -v -race -coverprofile=coverage.out ./...

# Integration tests (requires Firestore emulator)
export FIRESTORE_EMULATOR_HOST=localhost:8080
go test -tags=integration -v ./internal/storage/...
```

**Frontend:**
```bash
cd dataAnalysis
npm test
npm run test:coverage
```

### 4. Commit Changes

Use [Conventional Commits](https://www.conventionalcommits.org/) format:

```
<type>: <description>

[optional body]
```

**Types:**
- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation
- `style:` - Code style (formatting, etc.)
- `refactor:` - Code refactoring
- `test:` - Add or modify tests
- `ci:` - CI/CD changes

**Examples:**
```bash
git commit -m "feat(alerts-service): add pagination support"
git commit -m "fix: resolve data race in handler tests"
git commit -m "docs: update API documentation"
```

### 5. Push and Create PR

```bash
git push origin your-branch-name
```

Then create a pull request on GitHub targeting the `main` branch.

## Pull Request Requirements

### Required Checks

All PRs must pass these automated checks:

1. **Go Services** (if modified):
   - Linting (golangci-lint)
   - Unit tests with race detection
   - Coverage thresholds:
     - Alerts Service: 20%
     - Scraper Service: 20%
     - Archive Service: 20%
     - Internal packages: 15%

2. **Frontend** (if modified):
   - Tests
   - Coverage: 100% (strict)

3. **Terraform** (if modified):
   - Formatting check
   - Validation
   - Plan execution

4. **Integration Tests**:
   - Firebase emulator-based tests for storage layer

### PR Guidelines

- **Title**: Use conventional commit format
- **Description**: Explain what changed and why
- **Testing**: Describe how you tested the changes
- **Breaking Changes**: Clearly document any breaking changes

### Review Process

- PRs are reviewed before merging
- Address review feedback by pushing additional commits
- Once approved and checks pass, PRs can be merged

## CI/CD Pipeline

### On Pull Request

All relevant workflows run automatically:
- `Alerts Service CI/CD`: Test alerts service
- `Scraper Service CI/CD`: Test scraper service
- `Archive Service CI/CD`: Test archive service
- `Shared Internal Tests`: Unit and integration tests for shared packages
- `Frontend CI`: Lint and test frontend
- `Terraform CI/CD`: Plan Terraform changes

Deployment jobs are **skipped** for PRs.

### On Merge to Main

Same checks run, plus:
- Docker images are built and pushed to Artifact Registry
- Services are deployed to Cloud Run
- Terraform changes are applied (if approved)

## Code Style

### Go

- Use `gofmt` and `gofmt -s` for formatting
- Follow [Effective Go](https://golang.org/doc/effective_go)
- Write table-driven tests
- Use meaningful variable names
- Add comments for exported functions

### JavaScript

- Follow existing code patterns in `dataAnalysis/`
- Use ES6+ features
- Write tests for new functions
- Keep functions small and focused

### Terraform

- Use `terraform fmt` for formatting
- Follow module structure in `terraform/modules/`
- Document variables and outputs
- Use meaningful resource names

## Running Services Locally

### Backend Services

```bash
# Set environment variables
export $(cat .env | xargs)

# Run scraper
go run ./cmd/scraper-service/main.go

# Run alerts service
go run ./cmd/alerts-service/main.go

# Run archive service
go run ./cmd/archive-service/main.go
```

### Frontend

```bash
cd dataAnalysis
firebase emulators:start
```

Access at `http://localhost:5000`

## Testing Against Real GCP Resources

For integration testing with real GCP services:

1. Set up a test GCP project
2. Configure `.env` with test project details
3. Authenticate: `gcloud auth application-default login`
4. Run services against test environment

**Note**: Avoid testing against production resources.

## Project Structure

```
cmd/              # Service entry points
internal/         # Shared packages
  ├── auth/       # Authentication
  ├── models/     # Data models
  ├── storage/    # Firestore/GCS storage
  └── waze/       # Waze API client
dataAnalysis/     # Frontend dashboard
terraform/        # Infrastructure as Code
.github/workflows/ # CI/CD pipelines
```

## Getting Help

- Open an [issue](https://github.com/Lllllllleong/wazePoliceScraperGCP/issues) for bugs or questions
- Check existing issues and PRs first
- For security issues, see [SECURITY.md](SECURITY.md)

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
