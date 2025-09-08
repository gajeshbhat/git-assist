# Installation Guide

This guide covers different ways to install git-assist on various platforms.

## Quick Install (Recommended)

### macOS and Linux

```bash
curl -sSL https://raw.githubusercontent.com/gajeshbhat/git-assist/main/install.sh | bash
```

This script will:
- Detect your operating system and architecture
- Download the latest release
- Install to `/usr/local/bin` (or `~/bin` if no sudo access)
- Set up shell completion automatically

### Windows

Download the latest Windows release from [GitHub Releases](https://github.com/gajeshbhat/git-assist/releases) and add to your PATH.

## Manual Installation

### From GitHub Releases

1. Go to [GitHub Releases](https://github.com/gajeshbhat/git-assist/releases)
2. Download the appropriate binary for your platform:
   - `git-assist-linux-amd64.tar.gz` - Linux 64-bit
   - `git-assist-darwin-amd64.tar.gz` - macOS Intel
   - `git-assist-darwin-arm64.tar.gz` - macOS Apple Silicon
   - `git-assist-windows-amd64.zip` - Windows 64-bit

3. Extract and install:
   ```bash
   tar -xzf git-assist-*.tar.gz
   sudo mv git-assist /usr/local/bin/
   ```

### Package Managers

#### Homebrew (macOS/Linux)

```bash
brew tap gajeshbhat/git-assist
brew install git-assist
```

#### Arch Linux (AUR)

```bash
yay -S git-assist
```

#### Snap (Linux)

```bash
sudo snap install git-assist
```

## Build from Source

### Prerequisites

- Go 1.19 or later
- Git
- Make (optional, for convenience)

### Build Steps

1. Clone the repository:
   ```bash
   git clone https://github.com/gajeshbhat/git-assist.git
   cd git-assist
   ```

2. Build the binary:
   ```bash
   make build
   # or
   go build -o git-assist cmd/git-assist/main.go
   ```

3. Install:
   ```bash
   sudo mv git-assist /usr/local/bin/
   # or
   mv git-assist ~/bin/  # if ~/bin is in your PATH
   ```

### Development Build

For development with additional debugging:

```bash
make dev-build
```

## Post-Installation Setup

### 1. Verify Installation

```bash
git-assist --version
```

### 2. Setup AI Backend

Choose one of the following AI backends:

#### Option A: Ollama (Recommended - Local AI)

```bash
git-assist config --install-ollama
git-assist config --start-service
git-assist config --pull-model codellama:7b
git-assist config --set-model codellama:7b
```

#### Option B: OpenAI API

```bash
export OPENAI_API_KEY="your-api-key"
git-assist config --set-provider openai
git-assist config --set-model gpt-3.5-turbo
```

#### Option C: Anthropic Claude

```bash
export ANTHROPIC_API_KEY="your-api-key"
git-assist config --set-provider anthropic
git-assist config --set-model claude-3-sonnet
```

### 3. Setup Shell Completion

```bash
git-assist config --setup-completion
```

This will automatically detect your shell and install completion.

### 4. Test the Setup

```bash
git-assist config --test-ai
```

### 5. Initialize in Repository

```bash
cd your-git-repository
git-assist init
```

## Platform-Specific Notes

### macOS

- **Apple Silicon**: Use the `darwin-arm64` binary
- **Intel**: Use the `darwin-amd64` binary
- **Homebrew**: Recommended installation method
- **Gatekeeper**: You may need to allow the binary in System Preferences

### Linux

- **Distribution packages**: Available for major distributions
- **AppImage**: Portable version available
- **Permissions**: Ensure the binary is executable (`chmod +x git-assist`)

### Windows

- **PowerShell**: Recommended terminal
- **WSL**: Linux installation works in WSL
- **PATH**: Add installation directory to system PATH
- **Antivirus**: May need to whitelist the binary

## Troubleshooting

### Permission Denied

```bash
chmod +x git-assist
```

### Command Not Found

Ensure the installation directory is in your PATH:

```bash
echo $PATH
which git-assist
```

Add to PATH if needed:

```bash
# Add to ~/.bashrc, ~/.zshrc, etc.
export PATH="$PATH:/usr/local/bin"
```

### AI Backend Issues

```bash
# Test AI connection
git-assist config --test-ai

# Check configuration
git-assist config --show

# Reset configuration
git-assist config --reset
```

### Shell Completion Not Working

```bash
# Reinstall completion
git-assist config --setup-completion

# Manual setup (if automatic fails)
git-assist completion zsh > ~/.zsh/completions/_git-assist
```

## Uninstallation

### Remove Binary

```bash
sudo rm /usr/local/bin/git-assist
# or
rm ~/bin/git-assist
```

### Remove Configuration

```bash
rm -rf ~/.git-assist
```

### Remove Shell Completion

```bash
# zsh
rm ~/.zsh/completions/_git-assist

# bash
sudo rm /etc/bash_completion.d/git-assist

# fish
rm ~/.config/fish/completions/git-assist.fish
```

## Updating

### Automatic Update

```bash
git-assist update
```

### Manual Update

1. Download the latest release
2. Replace the existing binary
3. Restart your shell

### From Source

```bash
cd git-assist
git pull origin main
make build
sudo mv git-assist /usr/local/bin/
```

## Docker Usage

### Run in Container

```bash
docker run --rm -v $(pwd):/workspace gajeshbhat/git-assist:latest analyze
```

### Build Docker Image

```bash
docker build -t git-assist .
```

## Configuration

After installation, git-assist stores configuration in:

- **Config file**: `~/.git-assist/config.yaml`
- **Cache**: `~/.git-assist/cache/`
- **Logs**: `~/.git-assist/logs/`

See the [Configuration Guide](CONFIGURATION.md) for detailed setup options.
