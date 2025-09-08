#!/bin/bash

# Git Assist Test Setup Script
# This script sets up a test environment for git-assist

set -e

echo "🧪 Setting up git-assist test environment..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if we're in the git-assist project directory
if [ ! -f "go.mod" ] || ! grep -q "git-assist" go.mod; then
    print_error "Please run this script from the git-assist project root directory"
    exit 1
fi

# Build git-assist
print_status "Building git-assist..."
make build

# Install to user bin
print_status "Installing git-assist to ~/bin..."
mkdir -p ~/bin
cp git-assist ~/bin/

# Check if ~/bin is in PATH
if [[ ":$PATH:" != *":$HOME/bin:"* ]]; then
    print_warning "~/bin is not in your PATH"
    print_status "Adding ~/bin to PATH in shell profile..."

    # Detect shell and add to appropriate file
    if [ -n "$ZSH_VERSION" ]; then
        echo 'export PATH="$HOME/bin:$PATH"' >> ~/.zshrc
        print_success "Added to ~/.zshrc"
    elif [ -n "$BASH_VERSION" ]; then
        echo 'export PATH="$HOME/bin:$PATH"' >> ~/.bashrc
        print_success "Added to ~/.bashrc"
    else
        echo 'export PATH="$HOME/bin:$PATH"' >> ~/.profile
        print_success "Added to ~/.profile"
    fi
    print_status "Run 'source ~/.zshrc' (or appropriate file) or start a new terminal"
fi

# Create test repository
TEST_DIR="/tmp/git-assist-test-$(date +%s)"
print_status "Creating test repository at $TEST_DIR..."
mkdir -p "$TEST_DIR"
cd "$TEST_DIR"

# Initialize git repository
git init
git config user.name "Test User"
git config user.email "test@example.com"

# Initialize git-assist
export PATH="$HOME/bin:$PATH"
git-assist init

# Create some test files
print_status "Creating test files..."
cat > README.md << 'EOF'
# Test Project

This is a test project for git-assist.

## Features

- Feature 1
- Feature 2
- Feature 3
EOF

mkdir -p src
cat > src/main.go << 'EOF'
package main

import "fmt"

func main() {
    fmt.Println("Hello, git-assist!")
}
EOF

cat > .gitignore << 'EOF'
# Build artifacts
*.exe
*.dll
*.so
*.dylib

# Test binary
*.test

# Output of the go coverage tool
*.out

# Dependency directories
vendor/

# IDE files
.vscode/
.idea/
*.swp
*.swo
EOF

print_success "Test repository created at $TEST_DIR"
print_success "git-assist initialized in test repository"

echo ""
print_status "Test commands you can try:"
echo "  cd $TEST_DIR"
echo "  git add README.md"
echo "  git-assist commit --dry-run"
echo "  git-assist --help"

echo ""
print_status "To test in any new terminal:"
echo "  1. Open a new terminal"
echo "  2. cd $TEST_DIR"
echo "  3. git-assist --version"

echo ""
print_success "Setup complete! 🎉"
