# Configuration Reference

Complete reference for git-assist configuration options.

## Configuration Files

Git-assist uses a hierarchical configuration system:

1. **Global**: `~/.git-assist/config.json`
2. **Repository**: `.git/git-assist/config.json`
3. **Project**: `.git-assist-rules.json`

## Configuration Schema

```json
{
  "models": [
    {
      "type": "ollama",
      "name": "codellama:7b",
      "endpoint": "http://localhost:11434"
    }
  ],
  "practices": {
    "industry": "conventional",
    "custom_file": ".git-assist-rules.md",
    "rules": ["require-type", "max-length-50"]
  },
  "preferences": {
    "auto_stage": false,
    "explain_commands": true,
    "output_format": "text"
  }
}
```


*This documentation will be auto-generated from configuration structs.*
