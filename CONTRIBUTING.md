# Contributing to Scrapp'd

Thank you for your interest in contributing to Scrapp'd! This document provides guidelines and instructions for contributing to the project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Code Style Guidelines](#code-style-guidelines)
- [Testing Guidelines](#testing-guidelines)
- [Commit Guidelines](#commit-guidelines)
- [Pull Request Process](#pull-request-process)
- [Project Structure](#project-structure)

## Code of Conduct

We are committed to providing a welcoming and inclusive environment. Please be respectful and professional in all interactions.

## Getting Started

### Prerequisites

- Docker & Docker Compose
- Go 1.21+
- Python 3.10+
- Flutter 3.16+
- Git

### Setup Development Environment

1. Fork the repository
2. Clone your fork:
```bash
git clone https://github.com/your-username/scrappd-app.git
cd scrappd-app
```

3. Run the setup script:
```bash
make setup
```

4. Start the services:
```bash
make start-infra
make dev-api    # Terminal 1
make dev-ml     # Terminal 2
```

## Development Workflow

### Creating a New Feature

1. Create a new branch from `main`:
```bash
git checkout -b feature/your-feature-name
```

2. Make your changes following our [code style guidelines](#code-style-guidelines)

3. Write tests for your changes

4. Run tests locally:
```bash
make test
```

5. Run linters:
```bash
make lint
```

6. Commit your changes following our [commit guidelines](#commit-guidelines)

7. Push to your fork and create a pull request

### Branch Naming Convention

- `feature/` - New features
- `fix/` - Bug fixes
- `refactor/` - Code refactoring
- `docs/` - Documentation changes
- `test/` - Test improvements
- `chore/` - Maintenance tasks

Examples:
- `feature/add-canvas-zoom`
- `fix/image-upload-error`
- `refactor/auth-service`

## Code Style Guidelines

### Go (API Service)

- Follow [Effective Go](https://golang.org/doc/effective_go)
- Use `gofmt` for formatting
- Run `golangci-lint` before committing
- Write meaningful comments for exported functions
- Keep functions small and focused
- Use dependency injection
- Follow clean architecture principles

Example:
```go
// CreateScrapbook creates a new scrapbook for the user
func (s *ScrapbookService) CreateScrapbook(ctx context.Context, req *CreateScrapbookRequest) (*Scrapbook, error) {
    if err := s.validator.Validate(req); err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }
    
    // Implementation
}
```

### Python (ML Service)

- Follow [PEP 8](https://pep8.org/)
- Use `black` for formatting
- Use `isort` for import sorting
- Use type hints
- Write docstrings for all public functions
- Keep functions pure when possible

Example:
```python
def remove_background(
    image: np.ndarray,
    model: torch.nn.Module,
    threshold: float = 0.5
) -> np.ndarray:
    """Remove background from an image using the trained model.
    
    Args:
        image: Input image as numpy array
        model: Trained PyTorch model
        threshold: Confidence threshold for mask
        
    Returns:
        Image with background removed
        
    Raises:
        ValueError: If image is invalid
    """
    # Implementation
```

### Flutter (Mobile App)

- Follow [Effective Dart](https://dart.dev/guides/language/effective-dart)
- Use `dart format` for formatting
- Use named parameters for functions with multiple arguments
- Organize code by features
- Use BLoC pattern for state management
- Write widget tests

Example:
```dart
class ScrapbookCard extends StatelessWidget {
  const ScrapbookCard({
    Key? key,
    required this.scrapbook,
    this.onTap,
  }) : super(key: key);

  final Scrapbook scrapbook;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    // Implementation
  }
}
```

## Testing Guidelines

### Unit Tests

Write unit tests for all business logic:

```bash
# API tests
cd services/api
make test-unit

# ML service tests
cd services/ml-service
pytest tests/unit -v

# Mobile tests
cd mobile
flutter test
```

### Integration Tests

Write integration tests for API endpoints and workflows:

```bash
# API integration tests
cd services/api
make test-integration

# ML service integration tests
cd services/ml-service
pytest tests/integration -v
```

### Test Coverage

- Aim for >80% code coverage
- Focus on critical paths
- Test edge cases and error handling

## Commit Guidelines

We follow [Conventional Commits](https://www.conventionalcommits.org/) specification.

### Commit Message Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

### Examples

```
feat(api): add scrapbook sharing endpoint

Add POST /api/v1/scrapbooks/:id/share endpoint to allow users
to share their scrapbooks with specific users or make them public.

Closes #123
```

```
fix(ml): resolve memory leak in background removal

Fixed memory leak caused by not releasing GPU tensors after
processing. Added proper cleanup in the finally block.

Fixes #456
```

## Pull Request Process

1. **Update Documentation**: Ensure README and relevant docs are updated

2. **Add Tests**: All new features must include tests

3. **Run CI Checks**: Ensure all tests and lints pass
```bash
make ci
```

4. **Update CHANGELOG**: Add your changes to CHANGELOG.md

5. **Create PR**: 
   - Use a clear, descriptive title
   - Reference related issues
   - Describe what changed and why
   - Add screenshots for UI changes

6. **Review Process**:
   - Address review comments promptly
   - Keep commits clean and logical
   - Squash commits if needed

### PR Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] Manual testing completed

## Screenshots (if applicable)
Add screenshots here

## Related Issues
Closes #123
```

## Project Structure

### Key Directories

```
scrappd-app/
├── mobile/              # Flutter mobile app
├── services/
│   ├── api/            # Go backend API
│   └── ml-service/     # Python ML service
├── infrastructure/     # IaC and deployment
├── docs/              # Documentation
└── scripts/           # Utility scripts
```

### Adding New Features

#### API Endpoint

1. Define domain model in `internal/domain/`
2. Create repository in `internal/repository/`
3. Implement service in `internal/service/`
4. Add handler in `internal/handler/`
5. Register route in `internal/router/`
6. Write tests

#### ML Model

1. Add model class in `src/models/`
2. Implement preprocessing in `src/services/`
3. Create API endpoint in `src/api/routes/`
4. Add tests in `tests/`
5. Update documentation

#### Mobile Feature

1. Create feature directory in `lib/features/`
2. Implement data layer (models, repositories)
3. Create domain layer (entities, use cases)
4. Build presentation layer (BLoC, pages, widgets)
5. Write tests

## Questions?

If you have questions:
- Check existing issues and discussions
- Create a new issue with the `question` label
- Reach out to maintainers

Thank you for contributing to Scrapp'd! 🎨✨