#!/bin/bash

set -e

APP_NAME="argus"
INSTALL_DIR="$HOME/.local/bin"

echo "[*] Building $APP_NAME..."
GOOS=$(uname | tr '[:upper:]' '[:lower:]')
GOARCH=$(uname -m)

if [[ "$GOARCH" == "x86_64" ]]; then
  GOARCH="amd64"
elif [[ "$GOARCH" == "aarch64" || "$GOARCH" == "arm64" ]]; then
  GOARCH="arm64"
fi

go version | awk '{print $3}' | sed 's/go//;s/\.[0-9]*$//' | xargs -I{} sed -i "s/^go .*/go {}/" go.mod

GOOS=$GOOS GOARCH=$GOARCH go build -o "$APP_NAME" -ldflags="-s -w" .

echo "[*] Ensuring $INSTALL_DIR exists..."
mkdir -p "$INSTALL_DIR"

echo "[*] Installing to $INSTALL_DIR..."
mv "$APP_NAME" "$INSTALL_DIR/$APP_NAME"

if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
  echo "[!] $INSTALL_DIR is not in your PATH."
  echo "    You can add it by adding this to your shell config:"
  echo "    export PATH=\"$INSTALL_DIR:\$PATH\""
else
  echo -e "\n\n[✔] Installed! You can now run '$APP_NAME configure'/'$APP_NAME c'."
fi

cp -r ./config/* $HOME/.config/$APP_NAME/

# Set up completions

SHELL_NAME=$(basename "$SHELL")

case "$SHELL_NAME" in
    bash)
        FILE="$HOME/.bashrc"
        COMPLETION_LINE='source <(argus completion bash)'

        if ! grep -Fxq "$COMPLETION_LINE" "$FILE" 2>/dev/null; then
            echo -e "\n\n# SETUP ARGUS COMPLETION" >> "$FILE" 2>/dev/null
            echo "$COMPLETION_LINE" >> "$FILE" 2>/dev/null
        fi
        source "$FILE"
        ;;
    zsh)
        FILE="$HOME/.zshrc"
        COMPLETION_LINE='source <(argus completion zsh)'

        if ! grep -Fxq "$COMPLETION_LINE" "$FILE" 2>/dev/null; then
            echo -e "\n\n# SETUP ARGUS COMPLETION" >> "$FILE" 2>/dev/null
            echo "$COMPLETION_LINE" >> "$FILE" 2>/dev/null
        fi
        source "$FILE" >/dev/null 2>&1
        ;;
    fish)
        FILE="$HOME/.config/fish/completions/argus.fish"
        mkdir -p "$(dirname "$FILE")" >/dev/null 2>&1
        if ! grep -Fq "argus completion fish" "$FILE" 2>/dev/null; then
            argus completion fish > "$FILE" 2>/dev/null
        fi
        ;;
    *)
        exit 1
        ;;
esac
