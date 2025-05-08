#!/bin/bash

# JellyFaaS Installer Script
# This script installs JellyFaaS CLI on macOS and Linux by:
# 1. Checking for and optionally installing Go
# 2. Checking for and optionally installing Git
# 3. Cloning the JellyFaaS CLI repository
# 4. Building the binary
# 5. Installing it to /usr/local/bin

# Script version - change this when updating the script
SCRIPT_VERSION="1.0.2"

set -e

# Text formatting
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
BOLD='\033[1m'
NC='\033[0m' # No Color

# Detect OS
OS="$(uname -s)"
case "${OS}" in
    Linux*)     OS_TYPE="Linux";;
    Darwin*)    OS_TYPE="macOS";;
    *)          OS_TYPE="Unknown";;
esac

# Detect architecture
ARCH=$(uname -m)
if [ "$ARCH" = "arm64" ] || [ "$ARCH" = "aarch64" ]; then
    ARCH_TYPE="arm64"
    if [ "$OS_TYPE" = "Linux" ]; then
        ARCH_TYPE="arm64" # Linux arm64 is called aarch64 sometimes
    fi
elif [ "$ARCH" = "x86_64" ]; then
    ARCH_TYPE="amd64"
else
    ARCH_TYPE="$ARCH"
fi

# Spinner function
spinner() {
    local pid=$1
    local delay=0.1
    local spinstr='|/-\'
    while [ "$(ps -p $pid -o pid=)" ]; do
        local temp=${spinstr#?}
        printf " [%c]  " "$spinstr"
        local spinstr=$temp${spinstr%"$temp"}
        sleep $delay
        printf "\b\b\b\b\b\b"
    done
    printf "    \b\b\b\b"
}

# Function for tasks with spinner
run_with_spinner() {
    local message="$1"
    local command="$2"

    echo -ne "${BLUE}[INFO]${NC} $message"
    eval "$command" &
    spinner $!

    # Check if the command was successful
    if [ $? -eq 0 ]; then
        echo -e "\r${GREEN}[SUCCESS]${NC} $message"
        return 0
    else
        echo -e "\r${RED}[ERROR]${NC} $message"
        return 1
    fi
}

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

# Function to install Go on macOS
install_go_macos() {
    log_info "Preparing to install Go for macOS..."

    # Get latest stable version
    local GO_VERSION=$(curl -s https://go.dev/VERSION?m=text | head -n 1)
    log_info "Latest Go version is $GO_VERSION"

    local GO_PKG="$GO_VERSION.darwin-$ARCH_TYPE.pkg"
    local GO_URL="https://go.dev/dl/$GO_PKG"
    local GO_PKG_PATH="$TMP_DIR/$GO_PKG"

    if [ "$ARCH_TYPE" = "arm64" ]; then
        log_info "Detected Apple Silicon (M1/M2/M3) Mac"
    else
        log_info "Detected Intel Mac"
    fi

    log_info "Downloading Go installation package... (this may take a few minutes depending on your connection speed)"

    # Download with spinner
    (curl -s -L -o "$GO_PKG_PATH" "$GO_URL" > /dev/null 2>&1) &
    spinner $!

    if [ ! -f "$GO_PKG_PATH" ] || [ ! -s "$GO_PKG_PATH" ]; then
        log_error "Failed to download Go. Please check your internet connection and try again."
        exit 1
    fi
    log_success "Go downloaded successfully."

    log_info "Installing Go... (this may require your admin password and take a minute to complete)"
    (sudo installer -pkg "$GO_PKG_PATH" -target / > /dev/null 2>&1) &
    spinner $!

    # Check if Go is now available
    if ! command_exists go; then
        # Add Go to PATH for current session
        export PATH=$PATH:/usr/local/go/bin
        log_info "Added Go to PATH for current session"

        if ! command_exists go; then
            log_error "Go installation failed. Please try installing it manually."
            log_info "Please add Go to your PATH by adding the following line to your ~/.zshrc or ~/.bash_profile:"
            echo 'export PATH=$PATH:/usr/local/go/bin'

            # Try alternative path for Apple Silicon Macs
            if [ "$ARCH_TYPE" = "arm64" ]; then
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
    fi

    log_success "Go installed successfully."
    log_info "Using Go version: $(go version)"
}

# Function to install Go on Linux
install_go_linux() {
    log_info "Preparing to install Go for Linux..."

    # Get latest stable version
    local GO_VERSION=$(curl -s https://go.dev/VERSION?m=text | head -n 1)
    log_info "Latest Go version is $GO_VERSION"

    # Handle Linux-specific architecture naming
    local GO_LINUX_ARCH="$ARCH_TYPE"
    if [ "$ARCH_TYPE" = "arm64" ]; then
        GO_LINUX_ARCH="arm64"
    fi

    log_info "Detected Linux $GO_LINUX_ARCH architecture"

    local GO_TAR="$GO_VERSION.linux-$GO_LINUX_ARCH.tar.gz"
    local GO_URL="https://go.dev/dl/$GO_TAR"
    local GO_TAR_PATH="$TMP_DIR/$GO_TAR"

    log_info "Downloading Go archive... (this may take a few minutes depending on your connection speed)"

    # Download with spinner
    (curl -s -L -o "$GO_TAR_PATH" "$GO_URL" > /dev/null 2>&1) &
    spinner $!

    if [ ! -f "$GO_TAR_PATH" ] || [ ! -s "$GO_TAR_PATH" ]; then
        log_error "Failed to download Go. Please check your internet connection and try again."
        exit 1
    fi
    log_success "Go downloaded successfully."

    log_info "Installing Go... (this may require your admin password)"

    # Create destination directory if it doesn't exist
    if [ ! -d "/usr/local" ]; then
        sudo mkdir -p /usr/local
    fi

    # Remove any existing Go installation in /usr/local
    if [ -d "/usr/local/go" ]; then
        log_info "Removing existing Go installation..."
        sudo rm -rf /usr/local/go
    fi

    # Extract Go to /usr/local
    (sudo tar -C /usr/local -xzf "$GO_TAR_PATH" > /dev/null 2>&1) &
    spinner $!

    # Check if Go is now available
    if ! command_exists go; then
        # Add Go to PATH for current session
        export PATH=$PATH:/usr/local/go/bin
        log_info "Added Go to PATH for current session"

        if ! command_exists go; then
            log_error "Go installation failed. Please try installing it manually."
            log_info "Please add Go to your PATH by adding the following line to your ~/.bashrc or ~/.profile:"
            echo 'export PATH=$PATH:/usr/local/go/bin'
            exit 1
        fi
    fi

    # Suggest adding to PATH permanently
    log_info "To add Go to your PATH permanently, add this line to your ~/.bashrc or ~/.profile:"
    echo 'export PATH=$PATH:/usr/local/go/bin'

    log_success "Go installed successfully."
    log_info "Using Go version: $(go version)"
}

# Function to install Go based on OS
install_go() {
    if [ "$OS_TYPE" = "macOS" ]; then
        install_go_macos
    elif [ "$OS_TYPE" = "Linux" ]; then
        install_go_linux
    else
        log_error "Unsupported operating system: $OS_TYPE"
        exit 1
    fi
}

# Function to check and install Git on macOS
check_and_install_git_macos() {
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

# Function to check and install Git on Linux
check_and_install_git_linux() {
    if command_exists git; then
        log_success "Git is already installed."
        return 0
    fi

    log_info "Git is not installed."
    if confirm "Do you want to install Git?"; then
        log_info "Attempting to install Git..."

        # Try to detect package manager
        if command_exists apt-get; then
            log_info "Detected apt package manager. Installing Git..."
            (sudo apt-get update > /dev/null 2>&1 && sudo apt-get install -y git > /dev/null 2>&1) &
            spinner $!
        elif command_exists dnf; then
            log_info "Detected dnf package manager. Installing Git..."
            (sudo dnf install -y git > /dev/null 2>&1) &
            spinner $!
        elif command_exists yum; then
            log_info "Detected yum package manager. Installing Git..."
            (sudo yum install -y git > /dev/null 2>&1) &
            spinner $!
        elif command_exists pacman; then
            log_info "Detected pacman package manager. Installing Git..."
            (sudo pacman -Sy --noconfirm git > /dev/null 2>&1) &
            spinner $!
        elif command_exists zypper; then
            log_info "Detected zypper package manager. Installing Git..."
            (sudo zypper install -y git > /dev/null 2>&1) &
            spinner $!
        else
            log_error "Could not detect package manager. Please install Git manually and rerun this script."
            exit 1
        fi

        # Check again if git is available after installation
        if ! command_exists git; then
            log_error "Git installation failed. Please install Git manually."
            exit 1
        fi
        log_success "Git installed successfully."
    else
        log_error "Git is required but not installed. Installation aborted."
        exit 1
    fi
}

# Function to check and install Git based on OS
check_and_install_git() {
    if [ "$OS_TYPE" = "macOS" ]; then
        check_and_install_git_macos
    elif [ "$OS_TYPE" = "Linux" ]; then
        check_and_install_git_linux
    else
        log_error "Unsupported operating system: $OS_TYPE"
        exit 1
    fi
}

# Function to clone the repository
clone_repo() {
    log_info "Cloning JellyFaaS CLI repository..."
    # Use -q flag to silence git output and run with spinner
    (git clone -q https://github.com/Platform48/jellyfaas_cli.git "$TMP_DIR/jellyfaas_cli" > /dev/null 2>&1) &
    spinner $!

    if [ ! -d "$TMP_DIR/jellyfaas_cli" ]; then
        log_error "Failed to clone repository. Please check your internet connection."
        exit 1
    fi
    log_success "Repository cloned successfully."
}

# Function to build the binary
build_binary() {
    log_info "Building JellyFaaS CLI... (this may take a moment)"
    cd "$TMP_DIR/jellyfaas_cli"

    # Build with spinner
    (go build -o jellyfaas cmd/jellyfaas.go > /dev/null 2>&1) &
    spinner $!

    if [ ! -f "./jellyfaas" ]; then
        log_error "Failed to build JellyFaaS CLI."
        exit 1
    fi
    log_success "JellyFaaS CLI built successfully."
}

# Function to install the binary
install_binary() {
    # Get the new version before installation
    local NEW_VERSION=$(get_built_version)
    local INSTALL_DIR="/usr/local/bin"

    # Ensure install directory exists
    if [ ! -d "$INSTALL_DIR" ]; then
        log_info "Creating $INSTALL_DIR directory..."
        if ! sudo mkdir -p "$INSTALL_DIR"; then
            # Try alternative location for Linux if /usr/local/bin fails
            if [ "$OS_TYPE" = "Linux" ]; then
                INSTALL_DIR="$HOME/.local/bin"
                log_info "Trying alternative location: $INSTALL_DIR"
                mkdir -p "$INSTALL_DIR"
            else
                log_error "Failed to create $INSTALL_DIR directory."
                exit 1
            fi
        fi
    fi

    log_info "Installing JellyFaaS CLI to $INSTALL_DIR..."

    # Copy the binary
    if [ "$INSTALL_DIR" = "/usr/local/bin" ]; then
        # Use sudo for system directories
        if ! sudo cp "$TMP_DIR/jellyfaas_cli/jellyfaas" "$INSTALL_DIR/"; then
            log_error "Failed to copy binary to $INSTALL_DIR."
            exit 1
        fi

        if ! sudo chmod +x "$INSTALL_DIR/jellyfaas"; then
            log_error "Failed to make binary executable."
            exit 1
        fi
    else
        # Don't use sudo for user directories
        if ! cp "$TMP_DIR/jellyfaas_cli/jellyfaas" "$INSTALL_DIR/"; then
            log_error "Failed to copy binary to $INSTALL_DIR."
            exit 1
        fi

        if ! chmod +x "$INSTALL_DIR/jellyfaas"; then
            log_error "Failed to make binary executable."
            exit 1
        fi

        # If using user directory, ensure it's in PATH
        if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
            log_info "Adding $INSTALL_DIR to PATH for current session"
            export PATH="$PATH:$INSTALL_DIR"

            # Suggest adding to PATH permanently
            if [ -f "$HOME/.bash_profile" ]; then
                PROFILE_FILE="$HOME/.bash_profile"
            elif [ -f "$HOME/.profile" ]; then
                PROFILE_FILE="$HOME/.profile"
            elif [ -f "$HOME/.bashrc" ]; then
                PROFILE_FILE="$HOME/.bashrc"
            elif [ -f "$HOME/.zshrc" ]; then
                PROFILE_FILE="$HOME/.zshrc"
            else
                PROFILE_FILE="$HOME/.profile"
            fi

            log_info "To add JellyFaaS to your PATH permanently, add this line to $PROFILE_FILE:"
            echo "export PATH=\"\$PATH:$INSTALL_DIR\""
        fi
    fi

    # Verify installation and get the installed version
    if command_exists jellyfaas; then
        local INSTALLED_VERSION=$(jellyfaas version 2>/dev/null || echo "$NEW_VERSION")
        if [[ -n "$EXISTING_VERSION" && "$EXISTING_VERSION" != "unknown" ]]; then
            log_success "JellyFaaS CLI updated: $EXISTING_VERSION → $INSTALLED_VERSION"
        else
            log_success "JellyFaaS CLI version $INSTALLED_VERSION installed successfully to $INSTALL_DIR/jellyfaas"
        fi
    else
        log_success "JellyFaaS CLI installed successfully to $INSTALL_DIR/jellyfaas"
        if [ "$INSTALL_DIR" = "$HOME/.local/bin" ]; then
            log_info "You may need to restart your terminal or source your profile to use the 'jellyfaas' command."
        fi
    fi
}

# Main script execution
main() {
    echo -e "\n${BOLD}JellyFaaS CLI Installer for $OS_TYPE v${SCRIPT_VERSION}${NC}\n"

    # Check if OS is supported
    if [ "$OS_TYPE" = "Unknown" ]; then
        log_error "Unsupported operating system: $OS. This installer only supports macOS and Linux."
        exit 1
    fi

    log_info "Detected operating system: $OS_TYPE ($ARCH_TYPE)"

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