# Contributing to Eve

Thank you for considering contributing to Eve! This guide will help you understand how to contribute effectively to the project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [How Can I Contribute?](#how-can-i-contribute)
    - [Reporting Bugs](#reporting-bugs)
    - [Suggesting Features](#suggesting-features)
    - [Improving Documentation](#improving-documentation)
    - [Contributing Code](#contributing-code)
- [Development Workflow](#development-workflow)
- [Pull Request Guidelines](#pull-request-guidelines)

---

## Code of Conduct

By participating in this project, you agree to abide by the [Code of Conduct](CODE_OF_CONDUCT.md). Please treat everyone with respect and professionalism.

## How Can I Contribute?

### Reporting Bugs

If you encounter a bug, please open an issue with the following information:

1. **Description**: Provide a clear and concise description of the problem.
2. **Steps to Reproduce**: Include the steps to reproduce the issue.
3. **Expected Behavior**: Explain what you expected to happen.
4. **Actual Behavior**: Describe what actually happened.
5. **Environment**: Include details about your environment (e.g., Node.js version, operating system).

### Suggesting Features

We welcome feature suggestions! To suggest a feature:

1. Open an issue with the title prefixed by `[Feature Request]`.
2. Provide a detailed description of the feature and its benefits.
3. If possible, include examples or use cases.

### Improving Documentation

You can help by improving the documentation. This includes fixing typos, adding examples, or clarifying existing content. Feel free to open a pull request with your changes.

### Contributing Code

If you want to contribute code, please:

1. Fork the repository.
2. Create a new branch for your feature or bugfix.
3. Write clear, concise, and well-documented code.
4. Ensure your changes are covered by tests (if applicable).

## Development Workflow

1. Clone the repository:

    ```bash
    git clone https://github.com/wiibleyde/eve.git
    cd eve
    ```

2. Install dependencies:

    ```bash
    bun install
    ```

3. Create a `.env` file based on the [Installation section in the README](README.md#installation).

4. Run the bot locally:

    ```bash
    bun index.ts
    ```

## Pull Request Guidelines

1. Ensure your changes pass all tests and do not introduce any new issues.
2. Follow the existing coding style and conventions.
3. Write a clear and concise commit message.
4. Reference related issues in your pull request description (e.g., `Closes #123`).
5. Submit your pull request to the `main` branch.

---

Thank you for contributing to Eve! Your efforts help make this project better for everyone.
