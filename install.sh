#!/usr/bin/env bash
set -e

REPO="Yagnik-Gohil/img-opt"
INSTALL_DIR="/usr/local/bin"
BINARY="img-opt"

OS=$(uname | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

# Detect the correct binary
if [[ "$OS" == "darwin" && "$ARCH" == "arm64" ]]; then
  FILE="$BINARY-macos-arm64"
elif [[ "$OS" == "darwin" ]]; then
  FILE="$BINARY-macos"
elif [[ "$OS" == "linux" ]]; then
  FILE="$BINARY-linux"
elif [[ "$OS" == *"mingw"* || "$OS" == *"cygwin"* || "$OS" == *"msys"* ]]; then
  FILE="$BINARY.exe"
  INSTALL_DIR="/usr/bin"
else
  echo "❌ Unsupported OS/architecture: $OS/$ARCH"
  exit 1
fi

RAW_URL="https://raw.githubusercontent.com/$REPO/main/bin/$FILE"

echo "⬇️  Downloading $BINARY from $RAW_URL ..."
TMP_FILE=$(mktemp)

curl -fsSL "$RAW_URL" -o "$TMP_FILE"

echo "🚀 Installing to $INSTALL_DIR/$BINARY ..."
sudo mv "$TMP_FILE" "$INSTALL_DIR/$BINARY"
sudo chmod +x "$INSTALL_DIR/$BINARY"

echo "✅ $BINARY installed successfully!"
echo "Run '$BINARY --help' to get started."
