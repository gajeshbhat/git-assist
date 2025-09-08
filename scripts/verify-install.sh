#!/bin/bash

# Git Assist Installation Verification Script

echo "🔍 Verifying git-assist installation..."

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

success() {
    echo -e "${GREEN}✅ $1${NC}"
}

error() {
    echo -e "${RED}❌ $1${NC}"
}

warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

# Check if git-assist is in PATH
if command -v git-assist &> /dev/null; then
    success "git-assist found in PATH"
    echo "   Location: $(which git-assist)"
    
    # Test version
    VERSION=$(git-assist --version 2>/dev/null | head -n1)
    if [ $? -eq 0 ]; then
        success "Version check passed: $VERSION"
    else
        error "Version check failed"
    fi
    
    # Test help command
    if git-assist --help &> /dev/null; then
        success "Help command works"
    else
        error "Help command failed"
    fi
    
    # Test basic functionality
    if git-assist init --help &> /dev/null; then
        success "Init command works"
    else
        error "Init command failed"
    fi
    
else
    error "git-assist not found in PATH"
    echo ""
    echo "To fix this, run one of:"
    echo "  export PATH=\"\$HOME/bin:\$PATH\"  # Temporary fix"
    echo "  source ~/.zshrc                   # If already added to .zshrc"
    echo "  make install                      # Install to ~/bin"
    exit 1
fi

echo ""
echo "🎉 git-assist is ready to use!"
echo ""
echo "Try these commands:"
echo "  git-assist --help"
echo "  git-assist init  # (in a git repository)"
echo "  git-assist commit --dry-run  # (with staged changes)"
