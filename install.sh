#!/bin/bash

# JellyFaaS Installer Script
# This script installs JellyFaaS CLI on macOS by:
# 1. Checking for and optionally installing Go
# 2. Checking for and optionally installing Git
# 3. Cloning the JellyFaaS CLI repository
# 4. Building the binary
# 5. Installing it to /usr/local/bin

set -e

# Text formatting
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
BOLD='\033[1m'
NC='\033[0m' # No Color

# Temporary directory for the repository
TMP_DIR=$(mktemp -d 2>/dev/null)
trap 'rm -rf "$TMP_DIR" 2>/dev/null' EXIT

# Log functions
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

# Function to prompt for user confirmation
confirm() {
    local prompt="$1"
    local response

    while true; do
        read -p "$prompt [y/n/q]: " response
        case "$response" in
            [Yy]* ) return 0;;
            [Nn]* ) return 1;;
            [Qq]* ) log_info "Installation aborted."; exit 0;;
            * ) echo "Please answer yes (y), no (n), or quit (q).";;
        esac
    done
}

# Function to check if a command is available
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to check for existing JellyFaaS installation and version
check_existing_jellyfaas() {
    if command_exists jellyfaas; then
        log_info "JellyFaaS CLI is already installed."
        local EXISTING_VERSION
        EXISTING_VERSION=$(jellyfaas version 2>/dev/null || echo "unknown")
        log_info "Current version: $EXISTING_VERSION"
        return 0
    else
        return 1
    fi
}

# Function to get the version of the built binary
get_built_version() {
    local NEW_VERSION
    NEW_VERSION=$("$TMP_DIR/jellyfaas_cli/jellyfaas" version 2>/dev/null || echo "unknown")
    echo "$NEW_VERSION"
}

# Function to check Go version
check_go_version() {
    local version=$(go version | grep -oE 'go[0-9]+\.[0-9]+' | cut -c 3-)
    local major=$(echo "$version" | cut -d. -f1)
    local minor=$(echo "$version" | cut -d. -f2)

    # Require Go 1.19 or later
    if [ "$major" -gt 1 ] || ([ "$major" -eq 1 ] && [ "$minor" -ge 19 ]); then
        return 0
    else
        return 1
    fi
}

# Function to install Go
install_go() {
    log_info "Preparing to install Go..."

    # Get latest stable version
    local GO_VERSION=$(curl -s https://go.dev/VERSION?m=text | head -n 1)
    log_info "Latest Go version is $GO_VERSION"

    # Detect Mac architecture (Apple Silicon vs Intel)
    local ARCH=$(uname -m)
    local GO_PKG_ARCH="amd64"

    if [ "$ARCH" = "arm64" ]; then
        GO_PKG_ARCH="arm64"
        log_info "Detected Apple Silicon (M1/M2/M3) Mac"
    else
        log_info "Detected Intel Mac"
    fi

    local GO_PKG="$GO_VERSION.darwin-$GO_PKG_ARCH.pkg"
    local GO_URL="https://go.dev/dl/$GO_PKG"
    local GO_PKG_PATH="$TMP_DIR/$GO_PKG"

    log_info "Downloading Go installation package..."
    if ! curl -s -L -o "$GO_PKG_PATH" "$GO_URL"; then
        log_error "Failed to download Go. Please check your internet connection and try again."
        exit 1
    fi
    log_success "Go downloaded successfully."

    log_info "Installing Go... (this may require your admin password)"
    if ! sudo installer -pkg "$GO_PKG_PATH" -target /; then
        log_error "Failed to install Go. Please try installing it manually."
        exit 1
    fi
    log_success "Go installed successfully."

    # Add Go to PATH for current session
    export PATH=$PATH:/usr/local/go/bin
    log_info "Added Go to PATH for current session"

    # Check if Go is now available
    if ! command_exists go; then
        log_error "Go installation succeeded but executable not found in PATH."
        log_info "Please add Go to your PATH by adding the following line to your ~/.zshrc or ~/.bash_profile:"
        echo 'export PATH=$PATH:/usr/local/go/bin'

        # Try alternative path for Apple Silicon Macs
        if [ "$ARCH" = "arm64" ]; then
            log_info "Trying alternative path for Apple Silicon..."
            export PATH=$PATH:/opt/homebrew/bin:/opt/homebrew/go/bin

            if ! command_exists go; then
                log_error "Could not find Go in PATH. Please restart your terminal or add Go to your PATH manually."
                exit 1
            else
                log_success "Found Go in alternative path."
            fi
        else
            exit 1
        fi
    fi

    log_info "Using Go version: $(go version)"
}

# Function to check and install Git
check_and_install_git() {
    if command_exists git; then
        log_success "Git is already installed."
        return 0
    fi

    log_info "Git is not installed."
    if confirm "Do you want to install Git?"; then
        log_info "macOS will prompt you to install developer tools which includes Git."
        git --version

        # Check again if git is available after prompt
        if ! command_exists git; then
            log_error "Git installation failed or was cancelled. Please install Git manually."
            exit 1
        fi
        log_success "Git installed successfully."
    else
        log_error "Git is required but not installed. Installation aborted."
        exit 1
    fi
}

# Function to clone the repository
clone_repo() {
    log_info "Cloning JellyFaaS CLI repository..."
    # Use -q flag to silence git output and redirect any remaining output to /dev/null
    if ! git clone -q https://github.com/Platform48/jellyfaas_cli.git "$TMP_DIR/jellyfaas_cli" > /dev/null 2>&1; then
        log_error "Failed to clone repository. Please check your internet connection."
        exit 1
    fi
    log_success "Repository cloned successfully."
}

# Function to build the binary
build_binary() {
    log_info "Building JellyFaaS CLI..."
    cd "$TMP_DIR/jellyfaas_cli"
    if ! go build -o jellyfaas cmd/jellyfaas.go > /dev/null 2>&1; then
        log_error "Failed to build JellyFaaS CLI."
        exit 1
    fi
    log_success "JellyFaaS CLI built successfully."
}

# Function to install the binary
install_binary() {
    # Get the new version before installation
    local NEW_VERSION=$(get_built_version)

    log_info "Installing JellyFaaS CLI to /usr/local/bin..."
    if [ ! -d "/usr/local/bin" ]; then
        log_info "Creating /usr/local/bin directory..."
        if ! sudo mkdir -p /usr/local/bin; then
            log_error "Failed to create /usr/local/bin directory."
            exit 1
        fi
    fi

    if ! sudo cp "$TMP_DIR/jellyfaas_cli/jellyfaas" /usr/local/bin/; then
        log_error "Failed to copy binary to /usr/local/bin."
        exit 1
    fi

    if ! sudo chmod +x /usr/local/bin/jellyfaas; then
        log_error "Failed to make binary executable."
        exit 1
    fi

    # Verify installation and get the installed version
    if command_exists jellyfaas; then
        local INSTALLED_VERSION=$(jellyfaas version 2>/dev/null || echo "$NEW_VERSION")
        if [[ -n "$EXISTING_VERSION" && "$EXISTING_VERSION" != "unknown" ]]; then
            log_success "JellyFaaS CLI updated: $EXISTING_VERSION → $INSTALLED_VERSION"
        else
            log_success "JellyFaaS CLI version $INSTALLED_VERSION installed successfully to /usr/local/bin/jellyfaas"
        fi
    else
        log_success "JellyFaaS CLI installed successfully to /usr/local/bin/jellyfaas"
    fi
}

# Main script execution
main() {
    echo -e "\n${BOLD}JellyFaaS CLI Installer for macOS${NC}\n"

    # Check for existing JellyFaaS installation
    if check_existing_jellyfaas; then
        EXISTING_VERSION=$(jellyfaas version 2>/dev/null || echo "unknown")
    fi

    # Check for Go
    if command_exists go; then
        log_success "Go is already installed."
        if ! check_go_version; then
            log_warning "Your Go version may be outdated. JellyFaaS may require Go 1.19 or newer."
            if confirm "Do you want to update Go?"; then
                install_go
            else
                log_info "Continuing with existing Go installation."
            fi
        fi
    else
        log_info "Go is not installed."
        if confirm "Do you want to install Go?"; then
            install_go
        else
            log_error "Go is required but not installed. Installation aborted."
            exit 1
        fi
    fi

    # Check for Git
    check_and_install_git

    # Clone repository
    clone_repo

    # Build binary
    build_binary

    # Install binary
    install_binary

    echo -e "\n${BOLD}${GREEN}JellyFaaS CLI installation complete!${NC}"
    echo -e "You can now run the JellyFaaS CLI using the ${BOLD}jellyfaas${NC} command."
}

# Run the main function
main