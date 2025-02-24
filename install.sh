#!/bin/bash
set -e

# Define variables
GITHUB_REPO="foresturquhart/grimoire"
INSTALL_DIR="/usr/local/bin"
TMP_DIR="$(mktemp -d)"
LATEST_RELEASE_URL="https://api.github.com/repos/${GITHUB_REPO}/releases/latest"
USER_AGENT="Grimoire-Installer-Script"

# Detect OS and architecture
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

# Map architecture names
case "${ARCH}" in
  x86_64)
    ARCH="amd64"
    ;;
  aarch64|arm64)
    ARCH="arm64"
    ;;
  *)
    echo "Unsupported architecture: ${ARCH}"
    exit 1
    ;;
esac

# Handle OS-specific variations
case "${OS}" in
  darwin|linux|windows)
    # These operating systems are supported
    ;;
  *)
    echo "Unsupported operating system: ${OS}"
    exit 1
    ;;
esac

echo "Detected OS: ${OS}, Architecture: ${ARCH}"

# Fetch the latest release information
echo "Fetching latest release information..."
RELEASE_INFO=$(curl -s -H "User-Agent: ${USER_AGENT}" "${LATEST_RELEASE_URL}")
VERSION=$(echo "${RELEASE_INFO}" | grep -o '"tag_name": *"[^"]*"' | grep -o '[^"]*$')

# Construct download URL
if [ "${OS}" = "windows" ]; then
  ARCHIVE_FORMAT="zip"
else
  ARCHIVE_FORMAT="tar.gz"
fi

DOWNLOAD_URL="https://github.com/${GITHUB_REPO}/releases/download/${VERSION}/grimoire-${VERSION#v}-${OS}-${ARCH}.${ARCHIVE_FORMAT}"
echo "Downloading Grimoire ${VERSION} for ${OS}-${ARCH}..."

# Download the binary
curl -L -s -o "${TMP_DIR}/grimoire.${ARCHIVE_FORMAT}" "${DOWNLOAD_URL}"

# Extract the binary
echo "Extracting..."
cd "${TMP_DIR}"
if [ "${ARCHIVE_FORMAT}" = "zip" ]; then
  unzip -q "grimoire.${ARCHIVE_FORMAT}"
else
  tar -xzf "grimoire.${ARCHIVE_FORMAT}"
fi

# Find the binary in the extracted directory
BINARY_PATH=$(find . -type f -name "grimoire" -o -name "grimoire.exe" | head -n 1)

if [ -z "${BINARY_PATH}" ]; then
  echo "Error: Could not find grimoire binary in the downloaded archive."
  exit 1
fi

# If running as root, install to system directory
if [ "$(id -u)" -eq 0 ]; then
  # Install to system-wide directory
  echo "Installing Grimoire to ${INSTALL_DIR}..."
  cp "${BINARY_PATH}" "${INSTALL_DIR}/grimoire"
  chmod +x "${INSTALL_DIR}/grimoire"
else
  # Check if the user can write to the default location
  if [ -w "${INSTALL_DIR}" ]; then
    echo "Installing Grimoire to ${INSTALL_DIR}..."
    cp "${BINARY_PATH}" "${INSTALL_DIR}/grimoire"
    chmod +x "${INSTALL_DIR}/grimoire"
  else
    # Try to use sudo
    echo "Installing Grimoire to ${INSTALL_DIR} (requires sudo)..."
    sudo cp "${BINARY_PATH}" "${INSTALL_DIR}/grimoire"
    sudo chmod +x "${INSTALL_DIR}/grimoire"
  fi
fi

# Clean up
echo "Cleaning up..."
rm -rf "${TMP_DIR}"

# Verify installation
echo "Verifying installation..."
if command -v grimoire >/dev/null 2>&1; then
  echo "Grimoire ${VERSION} has been successfully installed!"
  echo "Run 'grimoire --help' to get started"
else
  echo "Installation successful, but grimoire is not in your PATH."
  echo "Make sure ${INSTALL_DIR} is in your PATH or manually move the binary to a directory in your PATH."
fi