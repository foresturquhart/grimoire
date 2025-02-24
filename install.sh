#!/bin/bash
set -e

# Define variables
GITHUB_REPO="foresturquhart/grimoire"
INSTALL_DIR="/usr/local/bin"
TMP_DIR="$(mktemp -d)"
LATEST_RELEASE_URL="https://api.github.com/repos/${GITHUB_REPO}/releases/latest"
USER_AGENT="Grimoire-Installer-Script"

cleanup() {
  echo "Cleaning up temporary files..."
  rm -rf "${TMP_DIR}"
}

# Set up trap to clean up on exit
trap cleanup EXIT

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
  darwin)
    OS="darwin"
    EXT="tar.gz"
    ;;
  linux)
    OS="linux"
    EXT="tar.gz"
    ;;
  windows*)
    OS="windows"
    EXT="zip"
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
VERSION=$(echo "${RELEASE_INFO}" | grep -o '"tag_name":"[^"]*"' | sed -E 's/"tag_name":"([^"]*)"/\1/')

if [ -z "${VERSION}" ]; then
  echo "Error: Could not extract version from release info."
  exit 1
fi

echo "Latest version: ${VERSION}"

# Create asset name pattern based on the naming convention in .goreleaser.yml
ASSET_NAME="grimoire-${VERSION#v}-${OS}-${ARCH}.${EXT}"
echo "Looking for asset: ${ASSET_NAME}"

# Extract the download URL for the specific asset
DOWNLOAD_URL=$(echo "${RELEASE_INFO}" | grep -o "\"browser_download_url\":\"[^\"]*${ASSET_NAME}\"" | sed -E 's/"browser_download_url":"([^"]*)"/\1/')

if [ -z "${DOWNLOAD_URL}" ]; then
  echo "Error: Could not find download URL for ${ASSET_NAME}."
  echo "Available assets:"
  echo "${RELEASE_INFO}" | grep -o '"browser_download_url":"[^"]*"' | sed -E 's/"browser_download_url":"([^"]*)"/\1/'
  exit 1
fi

echo "Download URL: ${DOWNLOAD_URL}"

# Create a temporary directory for downloading and extracting
echo "Downloading Grimoire ${VERSION} for ${OS}-${ARCH}..."
curl -L -o "${TMP_DIR}/${ASSET_NAME}" "${DOWNLOAD_URL}"

# Extract the archive
echo "Extracting archive..."
cd "${TMP_DIR}"
if [ "${EXT}" = "zip" ]; then
  unzip -q "${ASSET_NAME}"
else
  tar -xzf "${ASSET_NAME}"
fi

# Find the binary in the extracted directory
# According to .goreleaser.yml, it should be in a directory with the same name as the project
cd grimoire-*
BINARY_PATH="grimoire"
if [ "${OS}" = "windows" ]; then
  BINARY_PATH="grimoire.exe"
fi

if [ ! -f "${BINARY_PATH}" ]; then
  echo "Error: Could not find grimoire binary in the extracted archive."
  find . -type f | sort
  exit 1
fi

# Installation logic based on permissions
if [ "$(id -u)" -eq 0 ]; then
  # Running as root
  echo "Installing Grimoire to ${INSTALL_DIR}..."
  cp "${BINARY_PATH}" "${INSTALL_DIR}/grimoire"
  chmod +x "${INSTALL_DIR}/grimoire"
else
  # Not running as root, try to install or guide the user
  if [ -w "${INSTALL_DIR}" ]; then
    echo "Installing Grimoire to ${INSTALL_DIR}..."
    cp "${BINARY_PATH}" "${INSTALL_DIR}/grimoire"
    chmod +x "${INSTALL_DIR}/grimoire"
  else
    if command -v sudo >/dev/null 2>&1; then
      echo "Installing Grimoire to ${INSTALL_DIR} (requires sudo)..."
      sudo cp "${BINARY_PATH}" "${INSTALL_DIR}/grimoire"
      sudo chmod +x "${INSTALL_DIR}/grimoire"
    else
      # No sudo, offer to install to user's bin directory
      USER_BIN="${HOME}/.local/bin"
      echo "Cannot install to ${INSTALL_DIR} (permission denied and sudo not available)"
      
      if [ ! -d "${USER_BIN}" ]; then
        echo "Creating directory ${USER_BIN}..."
        mkdir -p "${USER_BIN}"
      fi
      
      echo "Installing to ${USER_BIN} instead..."
      cp "${BINARY_PATH}" "${USER_BIN}/grimoire"
      chmod +x "${USER_BIN}/grimoire"
      
      if [[ ":${PATH}:" != *":${USER_BIN}:"* ]]; then
        echo "Note: ${USER_BIN} is not in your PATH. You may need to add it:"
        echo "  export PATH=\"\$PATH:${USER_BIN}\""
        echo "Add this line to your shell's profile file (~/.bashrc, ~/.zshrc, etc.) to make it permanent."
      fi
    fi
  fi
fi

# Verify installation
echo "Verifying installation..."
GRIMOIRE_PATH=$(command -v grimoire 2>/dev/null || echo "")

if [ -n "${GRIMOIRE_PATH}" ]; then
  echo "Grimoire ${VERSION} has been successfully installed to ${GRIMOIRE_PATH}!"
  echo "Run 'grimoire --help' to get started"
else
  echo "Installation completed, but grimoire command is not in your PATH yet."
  
  # Check if we installed to a non-PATH location
  if [ -x "${HOME}/.local/bin/grimoire" ]; then
    echo "You can run it with: ${HOME}/.local/bin/grimoire"
    echo "Or add ${HOME}/.local/bin to your PATH to use 'grimoire' directly."
  elif [ -x "${INSTALL_DIR}/grimoire" ]; then
    echo "You can run it with: ${INSTALL_DIR}/grimoire"
    echo "Make sure ${INSTALL_DIR} is in your PATH to use 'grimoire' directly."
  fi
fi

echo "Installation complete!"