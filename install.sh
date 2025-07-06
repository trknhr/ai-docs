#!/bin/bash

set -e

REPO="trknhr/ai-docs"
BINARY_NAME="ai-docs"
INSTALL_DIR="/usr/local/bin"

# Detect OS and ARCH
OS="$(uname)"
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64) ARCH="x86_64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

# Fetch latest release tag
TAG=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
TAG_NO_V=$(echo "$TAG" | sed 's/^v//')

# Build download URL
FILENAME="${BINARY_NAME}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${TAG}/${FILENAME}"

# Download and extract
TMPDIR=$(mktemp -d)
echo "⬇️ Downloading $URL..."
curl -L "$URL" -o "$TMPDIR/$FILENAME"

echo "📦 Extracting..."
tar -xf "$TMPDIR/$FILENAME" -C "$TMPDIR"

echo "🚀 Installing to $INSTALL_DIR..."
chmod +x "$TMPDIR/$BINARY_NAME"
sudo mv "$TMPDIR/$BINARY_NAME" "$INSTALL_DIR/"

echo "✅ Installed: $INSTALL_DIR/$BINARY_NAME"