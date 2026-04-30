# Contributing to MLflow Kubernetes Operator

Thank you for your interest in contributing to the MLflow Kubernetes Operator! This document provides guidelines and instructions for contributing to the project.

## Code of Conduct

This project adheres to a Code of Conduct. By participating, you are expected to uphold this code. Please report unacceptable behavior to the project maintainers.

## Getting Started

### Prerequisites

- Go 1.21 or higher
- Docker (for building container images)
- Kubernetes cluster (for testing)
- kubectl configured to communicate with your cluster
- Helm 3.x (for Helm chart development)
- make (for build automation)

### Setting Up the Development Environment

1. **Fork and clone the repository**
   ```bash
   git clone https://github.com/your-username/mlflow-k8s-operator.git
   cd mlflow-k8s-operator
   ```

2. **Install dependencies**
   ```bash
   make deps
   ```

3. **Install development tools**
   ```bash
   make tools
   ```

4. **Build the operator**
   ```bash
   make build
   ```

## Development Workflow

### Making Changes

1. **Create a feature branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes** following the coding standards outlined below.

3. **Run tests**
   ```bash
   make test
   ```

4. **Format your code**
   ```bash
   make fmt
   ```

5. **Run linters**
   ```bash
   make vet
   ```

6. **Commit your changes**
   ```bash
   git commit -m "Add your commit message here"
   ```

7. **Push to your fork**
   ```bash
   git push origin feature/your-feature-name
   ```

8. **Create a Pull Request** to the main repository.

### Coding Standards

- Follow [Effective Go](https://golang.org/doc/effective_go) guidelines
- Use `gofmt` for code formatting
- Write meaningful commit messages following the [Conventional Commits](https://www.conventionalcommits.org/) specification
- Add tests for new functionality
- Update documentation as needed

### Pull Request Guidelines

- **Title**: Use a clear, descriptive title for your PR
- **Description**: Provide a detailed description of changes and the motivation behind them
- **Related Issues**: Link to any related issues using `Fixes #123` or `Closes #123`
- **Tests**: Ensure all tests pass
- **Documentation**: Update documentation if your changes affect user-facing behavior
- **Breaking Changes**: Clearly document any breaking changes

## Testing

### Running Unit Tests

```bash
make test
```

### Running E2E Tests

For E2E tests, you can use either envtest or a Kind cluster:

**Using envtest:**
```bash
cd test/e2e
go test -v
```

**Using Kind cluster:**
```bash
export USE_KIND=true
cd test/e2e
go test -v
```

### Test Coverage

To generate a test coverage report:
```bash
make test
go tool cover -html=cover.out
```

## Building and Deployment

### Building the Operator

```bash
make build
```

This creates the binary at `bin/manager`.

### Building Docker Image

```bash
make docker-build IMG=your-registry/mlflow-k8s-operator:tag
```

### Pushing Docker Image

```bash
make docker-push IMG=your-registry/mlflow-k8s-operator:tag
```

### Deploying to Kubernetes

**Using kustomize:**
```bash
make deploy IMG=your-registry/mlflow-k8s-operator:tag
```

**Using Helm:**
```bash
make helm-deploy
```

## Project Structure

```
mlflow-k8s-operator/
├── api/              # API definitions (CRDs)
│   └── v1alpha1/    # v1alpha1 API version
├── controllers/     # Controller implementations
├── config/          # Kubernetes manifests (kustomize)
├── charts/          # Helm charts
├── internal/        # Internal packages
├── test/            # Test files
│   └── e2e/        # End-to-end tests
├── main.go          # Main entry point
├── go.mod           # Go module file
├── Makefile         # Build automation
└── README.md        # Project documentation
```

## Adding New Features

### Adding a New API Type

1. Define the type in `api/v1alpha1/`
2. Run `make manifests` to generate CRDs
3. Implement the controller in `controllers/`
4. Add unit tests for the new type
5. Update documentation

### Adding Controller Logic

1. Implement the reconciliation logic in the appropriate controller
2. Add unit tests for the new logic
3. Test with a real Kubernetes cluster
4. Update the README if user-facing behavior changes

## Reporting Bugs

When reporting bugs, please include:

- A clear and descriptive title
- Steps to reproduce the issue
- Expected behavior
- Actual behavior
- Environment details (Kubernetes version, Go version, etc.)
- Relevant logs or error messages

Use the GitHub issue tracker to report bugs.

## Feature Requests

We welcome feature requests! Please:

- Check existing issues first to avoid duplicates
- Provide a clear description of the proposed feature
- Explain the use case and motivation
- Consider if you can contribute the implementation

## Documentation

Documentation is crucial for the project's success. When making changes:

- Update the README if user-facing behavior changes
- Add comments to complex code sections
- Update inline documentation in Go files
- Consider adding examples for new features

## Release Process

Releases are managed by the maintainers. The process includes:

1. Updating version numbers
2. Updating CHANGELOG.md
3. Creating a git tag
4. Building and releasing container images
5. Publishing Helm chart

## Questions?

Feel free to open an issue for questions or discussions. We're happy to help!

## License

By contributing to this project, you agree that your contributions will be licensed under the Apache License 2.0.
