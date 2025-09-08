# Git Assist Documentation

This documentation follows the [Diataxis framework](https://diataxis.fr/) for clear, purposeful documentation.

## Documentation Structure

### 📚 [Tutorials](./tutorials/) (Learning-oriented)
Step-by-step lessons for learning git-assist from scratch.

- [Getting Started](./tutorials/getting-started.md) - Your first steps with git-assist
- [Setting Up AI Models](./tutorials/setup-models.md) - Configure local AI models
- [Custom Rules](./tutorials/custom-rules.md) - Create team-specific git practices

### 🛠️ [How-to Guides](./how-to/) (Problem-oriented)
Practical solutions to specific problems.

- [Generate Better Commit Messages](./how-to/commit-messages.md)
- [Integrate with Existing Workflows](./how-to/workflow-integration.md)
- [Troubleshoot Common Issues](./how-to/troubleshooting.md)

### 📖 [Reference](./reference/) (Information-oriented)
Complete technical reference and API documentation.

- [Command Reference](./reference/commands.md) - All commands and flags
- [Configuration Reference](./reference/configuration.md) - All config options
- [API Documentation](./reference/api.md) - Generated from code comments

### 💡 [Explanation](./explanation/) (Understanding-oriented)
Concepts, design decisions, and background information.

- [Architecture Overview](./explanation/architecture.md) - How git-assist works
- [AI Model Integration](./explanation/ai-models.md) - Why and how we use AI
- [Design Decisions](./explanation/design-decisions.md) - Why we built it this way

## Auto-generated Documentation

### API Reference
```bash
# View package documentation
go doc github.com/gajeshbhat/git-assist/internal/git

# Start local documentation server
godoc -http=:6060
# Then visit: http://localhost:6060/pkg/github.com/gajeshbhat/git-assist/
```

### Command Help
```bash
# Built-in help system
git-assist --help
git-assist commit --help
```

## Contributing to Documentation

- **Code comments**: Use Go's documentation conventions for auto-generation
- **Tutorials**: Focus on learning outcomes, not just steps
- **How-to guides**: Solve specific, real-world problems
- **Reference**: Be complete and accurate
- **Explanation**: Provide context and reasoning
