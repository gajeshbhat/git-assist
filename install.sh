#!/bin/bash

# git-assist installation script
# Usage: curl -sSL https://raw.githubusercontent.com/gajeshbhat/git-assist/main/install.sh | bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
REPO="gajeshbhat/git-assist"
BINARY_NAME="git-assist"
INSTALL_DIR="/usr/local/bin"
USER_INSTALL_DIR="$HOME/bin"

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Detect OS and architecture
detect_platform() {
    local os arch
    
    case "$(uname -s)" in
        Darwin)
            os="darwin"
            ;;
        Linux)
            os="linux"
            ;;
        CYGWIN*|MINGW*|MSYS*)
            os="windows"
            ;;
        *)
            log_error "Unsupported operating system: $(uname -s)"
            exit 1
            ;;
    esac
    
    case "$(uname -m)" in
        x86_64|amd64)
            arch="amd64"
            ;;
        arm64|aarch64)
            arch="arm64"
            ;;
        armv7l)
            arch="arm"
            ;;
        *)
            log_error "Unsupported architecture: $(uname -m)"
            exit 1
            ;;
    esac
    
    echo "${os}-${arch}"
}

# Get latest release version
get_latest_version() {
    local version
    version=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    
    if [ -z "$version" ]; then
        log_error "Failed to get latest version"
        exit 1
    fi
    
    echo "$version"
}

# Download and install binary
install_binary() {
    local platform version download_url temp_dir
    
    platform=$(detect_platform)
    version=$(get_latest_version)
    
    log_info "Detected platform: $platform"
    log_info "Latest version: $version"
    
    # Create temporary directory
    temp_dir=$(mktemp -d)
    cd "$temp_dir"
    
    # Construct download URL
    if [[ "$platform" == "windows"* ]]; then
        download_url="https://github.com/${REPO}/releases/download/${version}/${BINARY_NAME}-${platform}.zip"
    else
        download_url="https://github.com/${REPO}/releases/download/${version}/${BINARY_NAME}-${platform}.tar.gz"
    fi
    
    log_info "Downloading from: $download_url"
    
    # Download
    if command -v curl >/dev/null 2>&1; then
        curl -sL "$download_url" -o "archive"
    elif command -v wget >/dev/null 2>&1; then
        wget -q "$download_url" -O "archive"
    else
        log_error "Neither curl nor wget is available"
        exit 1
    fi
    
    # Extract
    if [[ "$platform" == "windows"* ]]; then
        unzip -q archive
    else
        tar -xzf archive
    fi
    
    # Find binary
    local binary_path
    if [[ "$platform" == "windows"* ]]; then
        binary_path="${BINARY_NAME}.exe"
    else
        binary_path="$BINARY_NAME"
    fi
    
    if [ ! -f "$binary_path" ]; then
        log_error "Binary not found in archive"
        exit 1
    fi
    
    # Make executable
    chmod +x "$binary_path"
    
    # Install
    local install_path
    if [ -w "$INSTALL_DIR" ] || [ "$(id -u)" = "0" ]; then
        install_path="$INSTALL_DIR/$BINARY_NAME"
        log_info "Installing to $install_path"
        
        if [ "$(id -u)" = "0" ]; then
            cp "$binary_path" "$install_path"
        else
            sudo cp "$binary_path" "$install_path"
        fi
    else
        # Try user install directory
        mkdir -p "$USER_INSTALL_DIR"
        install_path="$USER_INSTALL_DIR/$BINARY_NAME"
        log_info "Installing to $install_path"
        cp "$binary_path" "$install_path"
        
        # Check if user bin is in PATH
        if [[ ":$PATH:" != *":$USER_INSTALL_DIR:"* ]]; then
            log_warning "Add $USER_INSTALL_DIR to your PATH:"
            echo "  export PATH=\"\$PATH:$USER_INSTALL_DIR\""
        fi
    fi
    
    # Cleanup
    cd /
    rm -rf "$temp_dir"
    
    log_success "git-assist installed successfully!"
}

# Setup shell completion
setup_completion() {
    log_info "Setting up shell completion..."
    
    if command -v "$BINARY_NAME" >/dev/null 2>&1; then
        if "$BINARY_NAME" config --setup-completion >/dev/null 2>&1; then
            log_success "Shell completion setup complete"
        else
            log_warning "Shell completion setup failed (you can run 'git-assist config --setup-completion' later)"
        fi
    else
        log_warning "Binary not in PATH, skipping completion setup"
    fi
}

# Verify installation
verify_installation() {
    log_info "Verifying installation..."
    
    if command -v "$BINARY_NAME" >/dev/null 2>&1; then
        local version
        version=$("$BINARY_NAME" --version 2>/dev/null | head -n1 || echo "unknown")
        log_success "Installation verified: $version"
        return 0
    else
        log_error "Installation verification failed"
        return 1
    fi
}

# Show next steps
show_next_steps() {
    echo
    log_info "Next steps:"
    echo "  1. Setup AI backend:"
    echo "     git-assist config --install-ollama"
    echo "     git-assist config --start-service"
    echo "     git-assist config --pull-model codellama:7b"
    echo
    echo "  2. Initialize in your repository:"
    echo "     cd your-git-repository"
    echo "     git-assist init"
    echo
    echo "  3. Start using:"
    echo "     git-assist commit"
    echo "     git-assist analyze"
    echo
    echo "  For help: git-assist --help"
    echo "  Documentation: https://github.com/${REPO}#readme"
}

# Main installation flow
main() {
    echo "git-assist installer"
    echo "==================="
    echo
    
    # Check dependencies
    if ! command -v curl >/dev/null 2>&1 && ! command -v wget >/dev/null 2>&1; then
        log_error "curl or wget is required"
        exit 1
    fi
    
    # Install binary
    install_binary
    
    # Verify installation
    if verify_installation; then
        # Setup completion
        setup_completion
        
        # Show next steps
        show_next_steps
    else
        log_error "Installation failed"
        exit 1
    fi
}

# Run main function
main "$@"
