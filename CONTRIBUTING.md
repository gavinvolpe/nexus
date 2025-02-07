# Contributing to Nexus

Thank you for your interest in contributing to Nexus! This document provides guidelines and instructions for contributing to the project.

## Table of Contents
- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Coding Standards](#coding-standards)
- [Documentation](#documentation)
- [Testing](#testing)
- [Pull Request Process](#pull-request-process)

## Code of Conduct

By participating in this project, you agree to maintain a respectful and inclusive environment. Key points:
- Use welcoming and inclusive language
- Be respectful of differing viewpoints and experiences
- Accept constructive criticism gracefully
- Focus on what's best for the community

## Getting Started

1. Fork the repository
2. Clone your fork:
   ```bash
   git clone https://github.com/your-username/nexus.git
   cd nexus
   ```
3. Add the upstream remote:
   ```bash
   git remote add upstream https://github.com/gavinvolpe/nexus.git
   ```
4. Create a new branch:
   ```bash
   git checkout -b feature/your-feature-name
   ```

## Development Workflow

1. **Keep Your Fork Updated**
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. **Make Your Changes**
   - Write clean, maintainable code
   - Follow the project's coding standards
   - Keep commits focused and atomic
   - Write meaningful commit messages following [Conventional Commits](https://www.conventionalcommits.org/)

3. **Document Your Changes**
   - Update README.md if needed
   - Update COMPONENTS.md for architectural changes
   - Add entries to NOTES.md for significant changes
   - Update API documentation

4. **Test Your Changes**
   - Write unit tests for new functionality
   - Write integration tests for component interactions
   - Ensure all existing tests pass
   - Add benchmarks for performance-critical code

## Coding Standards

### Go Code Style
- Follow the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` to format your code
- Follow package naming conventions
- Keep functions focused and small
- Use meaningful variable names
- Add comments for complex logic

### Project Structure
```
nexus/
├── internal/          # Internal packages
├── pkg/              # Public packages
├── prompts/          # Prompt system
├── docs/             # Documentation
└── examples/         # Example code
```

### Commit Messages
Follow the Conventional Commits specification:
```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

Types:
- feat: New feature
- fix: Bug fix
- docs: Documentation changes
- style: Code style changes
- refactor: Code refactoring
- test: Adding tests
- chore: Maintenance tasks

## Documentation

1. **Code Documentation**
   - Add godoc comments to all exported types and functions
   - Include examples for complex functionality
   - Document any assumptions or limitations

2. **Project Documentation**
   - README.md: Project overview and quick start
   - COMPONENTS.md: Architecture and design
   - NOTES.md: Change log and decisions
   - API documentation in code

3. **Examples**
   - Add examples for new features
   - Keep examples up-to-date
   - Include both simple and complex use cases

## Testing

1. **Unit Tests**
   - Write tests for all new functionality
   - Use table-driven tests where appropriate
   - Aim for high test coverage
   - Use meaningful test names

2. **Integration Tests**
   - Test component interactions
   - Test with real dependencies
   - Include error cases

3. **Benchmarks**
   - Add benchmarks for performance-critical code
   - Compare before/after for optimizations

## Pull Request Process

1. **Before Submitting**
   - Update documentation
   - Add/update tests
   - Run all tests locally
   - Format code with `gofmt`
   - Update dependencies if needed

2. **Pull Request Content**
   - Use the PR template
   - Describe your changes
   - Link related issues
   - List breaking changes
   - Include testing instructions

3. **Review Process**
   - Address review comments
   - Keep the PR focused
   - Be responsive to feedback
   - Update your branch when needed

4. **After Merge**
   - Delete your branch
   - Update your fork
   - Celebrate your contribution! 🎉

## Questions?

Feel free to:
- Open an issue for questions
- Join our community discussions
- Reach out to maintainers

Thank you for contributing to Nexus! 🙏