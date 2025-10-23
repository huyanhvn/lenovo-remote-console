# Contributing to Lenovo Remote Console

First off, thank you for considering contributing to Lenovo Remote Console! It's people like you that make this tool better for everyone.

## Code of Conduct

This project and everyone participating in it is governed by our Code of Conduct. By participating, you are expected to uphold this code. Please be respectful and constructive in all interactions.

## How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please check existing issues to avoid duplicates. When you create a bug report, please include as many details as possible:

* **Use a clear and descriptive title** for the issue
* **Describe the exact steps to reproduce the problem**
* **Provide specific examples** to demonstrate the steps
* **Describe the behavior you observed** and explain why it's a problem
* **Explain which behavior you expected to see** instead
* **Include screenshots** if applicable
* **Include your environment details**:
  * OS and version
  * Go version (`go version`)
  * Browser type and version
  * BMC/XCC firmware version if known

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. When creating an enhancement suggestion, please include:

* **Use a clear and descriptive title**
* **Provide a detailed description** of the suggested enhancement
* **Provide specific examples** to demonstrate how it would work
* **Describe the current behavior** and explain how your suggestion improves it
* **Explain why this enhancement would be useful** to most users

### Pull Requests

1. Fork the repo and create your branch from `main`
2. If you've added code that should be tested, add tests
3. Ensure the test suite passes (`go test ./...`)
4. Make sure your code follows Go best practices and conventions
5. Run `go fmt ./...` to format your code
6. Run `go vet ./...` to check for common issues
7. Update documentation as needed
8. Issue the pull request

## Development Setup

1. Clone the repository:
```bash
git clone https://github.com/huyanhvn/lenovo-remote-console.git
cd lenovo-remote-console
```

2. Install dependencies:
```bash
go mod download
```

3. Generate test certificates (for local development):
```bash
openssl req -x509 -newkey rsa:4096 -keyout server.key -out server.crt \
  -days 365 -nodes -subj "/CN=localhost"
```

4. Run tests:
```bash
go test ./...
```

5. Build the project:
```bash
go build -o lenovo-console cmd/lenovo-console/main.go
```

## Coding Style

* Follow standard Go conventions and idioms
* Use meaningful variable and function names
* Add comments for exported functions and types
* Keep functions small and focused
* Handle errors appropriately
* Use `gofmt` to format your code
* Run `golint` and address any issues

## Testing

* Write unit tests for new functionality
* Ensure all tests pass before submitting PR
* Aim for good test coverage
* Include both positive and negative test cases

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...
```

## Documentation

* Update the README.md if you change functionality
* Add godoc comments to all exported types and functions
* Include examples in documentation where helpful
* Update the examples/ directory if adding new features

## Commit Messages

* Use the present tense ("Add feature" not "Added feature")
* Use the imperative mood ("Move cursor to..." not "Moves cursor to...")
* Limit the first line to 72 characters or less
* Reference issues and pull requests liberally after the first line
* Consider starting the commit message with an applicable emoji:
  * 🎨 `:art:` when improving the format/structure of the code
  * 🐛 `:bug:` when fixing a bug
  * ✨ `:sparkles:` when introducing new features
  * 📝 `:memo:` when writing docs
  * 🔧 `:wrench:` when changing configuration files
  * ✅ `:white_check_mark:` when adding tests
  * 🔒 `:lock:` when dealing with security
  * ⬆️ `:arrow_up:` when upgrading dependencies
  * ♻️ `:recycle:` when refactoring code

## Project Structure

```
lenovo-remote-console/
├── cmd/
│   └── lenovo-console/    # CLI application
│       └── main.go
├── lenovoconsole/          # Main package
│   ├── console.go          # Core console functionality
│   └── template.go         # HTML template
├── examples/               # Usage examples
│   └── multiple_consoles.go
├── main.go                 # Backward compatibility wrapper
├── go.mod                  # Module definition
├── README.md               # Project documentation
├── LICENSE                 # MIT License
└── CONTRIBUTING.md         # This file
```

## Questions?

Feel free to open an issue with your question or reach out to the maintainers.

Thank you for contributing! 🎉
