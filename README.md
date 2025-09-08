# git-assist

AI-powered Git assistant that enhances your development workflow with intelligent commit messages, repository analysis, and smart git operations.

## Features

- **AI-Generated Commit Messages**: Context-aware commit messages based on your staged changes
- **Repository Analysis**: Comprehensive analysis of code structure, health, and patterns
- **Intelligent Branch Management**: Safe branch operations with cleanup and validation
- **Smart History Navigation**: Search and analyze commit history with AI insights
- **Safe Rebasing**: AI-guided rebasing with safety checks and conflict resolution
- **Shell Autocompletion**: Tab completion for all commands and options

## Installation

### Quick Install

```bash
# Download and install the latest release
curl -sSL https://raw.githubusercontent.com/gajeshbhat/git-assist/main/install.sh | bash
```

### Manual Installation

1. Download the latest release from [GitHub Releases](https://github.com/gajeshbhat/git-assist/releases)
2. Extract and move to your PATH:
   ```bash
   tar -xzf git-assist-*.tar.gz
   sudo mv git-assist /usr/local/bin/
   ```

### Build from Source

```bash
git clone https://github.com/gajeshbhat/git-assist.git
cd git-assist
make build
sudo mv git-assist /usr/local/bin/
```

## Quick Start

### 1. Setup AI Model

```bash
# Install and configure Ollama (recommended)
git-assist config --install-ollama
git-assist config --start-service
git-assist config --pull-model codellama:7b
git-assist config --set-model codellama:7b

# Test the setup
git-assist config --test-ai
```

### 2. Setup Shell Completion

```bash
# Automatic setup for your shell
git-assist config --setup-completion
```

### 3. Initialize in Repository

```bash
cd your-git-repository
git-assist init
```

### 4. Start Using

```bash
# Generate AI commit messages
git add .
git-assist commit

# Analyze your repository
git-assist analyze

# Manage branches intelligently
git-assist branch analyze
git-assist branch cleanup

# Navigate history with AI insights
git-assist history --search "authentication"
```

## Commands

### Core Commands

#### `git-assist commit`
Generate AI-powered commit messages based on staged changes.

```bash
git-assist commit                     # Interactive commit with AI suggestions
git-assist commit -m "custom message" # Use custom message with AI enhancement
git-assist commit --no-ai             # Disable AI for this commit
git-assist commit --with-context      # Include repository context
```

#### `git-assist analyze`
Analyze repository structure, health, and patterns.

```bash
git-assist analyze                    # Quick repository overview
git-assist analyze --structure        # Detailed code organization analysis
git-assist analyze --workflow         # Git workflow patterns
git-assist analyze --dependencies     # Dependency analysis
git-assist analyze --health           # Repository health metrics
git-assist analyze --all              # Complete analysis
```

#### `git-assist branch`
Intelligent branch management and analysis.

```bash
git-assist branch analyze             # Analyze branch relationships
git-assist branch cleanup             # Safe cleanup of merged branches
git-assist branch suggest             # Suggest branch names
git-assist branch --create feature-x  # Create branch with validation
git-assist branch --strategy          # Suggest merge strategy
```

#### `git-assist history`
Navigate and search git history with AI insights.

```bash
git-assist history                    # Enhanced history view
git-assist history --search "auth"    # Search commits by content/message
git-assist history --author john      # Filter by author
git-assist history --timeline         # Timeline view
git-assist history --feature login    # Track feature development
git-assist history --explain          # AI development story
```

#### `git-assist rebase`
AI-powered intelligent rebasing with safety checks.

```bash
git-assist rebase main               # Safe rebase with checks
git-assist rebase --interactive      # AI-guided interactive rebase
git-assist rebase --explain          # Learn about rebasing
git-assist rebase --conflicts        # Help resolve conflicts
git-assist rebase --cleanup          # Clean up commit history
```

### Configuration Commands

#### `git-assist config`
Configure AI models, settings, and preferences.

```bash
git-assist config                     # Show current configuration
git-assist config --test-ai           # Test AI connection
git-assist config --list-models       # List available models
git-assist config --set-model llama2  # Set AI model
git-assist config --setup-completion  # Setup shell completion
```

#### `git-assist init`
Initialize git-assist in a repository.

```bash
git-assist init                       # Setup with default settings
git-assist init --with-rules          # Include custom commit rules
git-assist init --index-repo          # Create repository index
```

## Configuration

### AI Models

git-assist supports multiple AI backends:

- **Ollama** (Recommended): Local AI models
- **OpenAI**: GPT models via API
- **Anthropic**: Claude models via API
- **Custom**: Any OpenAI-compatible API

### Configuration File

Configuration is stored in `~/.git-assist/config.yaml`:

```yaml
ai:
  provider: "ollama"
  model: "codellama:7b"
  endpoint: "http://localhost:11434"

commit:
  auto_stage: false
  include_context: true
  max_length: 72

analysis:
  auto_index: true
  include_dependencies: true
```

### Environment Variables

```bash
export OLLAMA_ENDPOINT="http://localhost:11434"
export OPENAI_API_KEY="your-api-key"
export ANTHROPIC_API_KEY="your-api-key"
```

## Examples

### Daily Workflow

```bash
# Morning: Check repository status
git-assist analyze

# Development: Create feature branch
git-assist branch --create feature/user-auth

# Commit work with AI messages
git add .
git-assist commit

# Before merging: Clean up history
git-assist rebase --interactive

# Cleanup: Remove merged branches
git-assist branch cleanup
```

### Code Review

```bash
# Understand recent changes
git-assist history --count 10

# Analyze specific feature
git-assist history --feature authentication

# Check branch differences
git-assist branch analyze
```

### Learning Git

```bash
# Learn about rebasing
git-assist rebase --explain

# Understand repository structure
git-assist analyze --all

# Get help with conflicts
git-assist rebase --conflicts
```

## Advanced Usage

### Custom Commit Rules

Create `.git-assist/rules.yaml` in your repository:

```yaml
commit:
  patterns:
    - pattern: "^(feat|fix|docs|style|refactor|test|chore):"
      description: "Use conventional commit format"
    - pattern: "^.{1,72}$"
      description: "Keep subject line under 72 characters"

  templates:
    - type: "feature"
      template: "feat({{scope}}): {{description}}"
    - type: "bugfix"
      template: "fix({{scope}}): {{description}}"
```

### Repository Indexing

Enable automatic repository indexing for better AI context:

```bash
git-assist config --index-repo
git-assist analyze --show-index
```

### Integration with Git Hooks

Add to `.git/hooks/prepare-commit-msg`:

```bash
#!/bin/sh
if [ -z "$2" ]; then
    git-assist commit --generate-only > "$1"
fi
```

## Troubleshooting

### AI Not Working

```bash
# Check AI configuration
git-assist config --test-ai

# Verify model is available
git-assist config --list-installed

# Restart AI service
git-assist config --start-service
```

### Performance Issues

```bash
# Rebuild repository index
git-assist init --index-repo

# Check repository health
git-assist analyze --health
```

### Shell Completion Not Working

```bash
# Reinstall completion
git-assist config --setup-completion

# Manual setup (if automatic fails)
git-assist completion zsh > ~/.zsh/completions/_git-assist
```

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Commit changes: `git commit -m 'Add amazing feature'`
4. Push to branch: `git push origin feature/amazing-feature`
5. Open a Pull Request

### Development Setup

```bash
git clone https://github.com/gajeshbhat/git-assist.git
cd git-assist
make deps
make test-all
make build
```

### Running Tests

```bash
# Run unit tests
make test

# Run tests with coverage
make test-coverage

# Run integration tests
make test-integration

# Run all tests
make test-all

# Run tests with race detection
make test-race
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

- **Issues**: [GitHub Issues](https://github.com/gajeshbhat/git-assist/issues)
- **Discussions**: [GitHub Discussions](https://github.com/gajeshbhat/git-assist/discussions)
- **Documentation**: [Wiki](https://github.com/gajeshbhat/git-assist/wiki)
