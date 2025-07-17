#!/bin/bash

BIN="argus"
BIN_PATH=$(which $BIN 2>/dev/null || true)

echo "[*] Starting Argus uninstallation process..."

if [[ -n "$BIN_PATH" && -f "$BIN_PATH" ]]; then
    echo "[*] Uninstalling $BIN from $BIN_PATH..."
    rm -f "$BIN_PATH"
    if [[ $? -eq 0 ]]; then
        echo "[✔] $BIN binary removed."
    else
        echo "[✖] Failed to remove $BIN binary. Check permissions."
    fi
else
    echo "[!] $BIN not found in PATH or already removed. Skipping binary removal."
fi

rm -rf $HOME/.config/argus

echo "[*] Removing Argus shell completion setup..."

# --- For Bash ---
BASHRC="$HOME/.bashrc"
COMPLETION_LINE_BASH='source <(argus completion bash)'
COMMENT_LINE_BASH='# SETUP ARGUS COMPLETION'

if [[ -f "$BASHRC" ]]; then
    if grep -Fxq "$COMPLETION_LINE_BASH" "$BASHRC" 2>/dev/null; then
        echo "    [*] Removing completion from $BASHRC..."
        sed -i.bak "/^${COMMENT_LINE_BASH}$/d; /^${COMPLETION_LINE_BASH}$/d" "$BASHRC" 2>/dev/null
        sed -i.bak -E '/^[[:space:]]*$/N;/\n[[:space:]]*$/D' "$BASHRC" 2>/dev/null
        rm -f "${BASHRC}.bak" 2>/dev/null
        echo "    [✔] Completion removed from $BASHRC."
    else
        echo "    [i] Argus completion line not found in $BASHRC. Skipping."
    fi
else
    echo "    [i] $BASHRC not found. Skipping Bash completion removal."
fi

# --- For Zsh ---
ZSHRC="$HOME/.zshrc"
COMPLETION_LINE_ZSH='source <(argus completion zsh)'
COMMENT_LINE_ZSH='# SETUP ARGUS COMPLETION'

if [[ -f "$ZSHRC" ]]; then
    if grep -Fxq "$COMPLETION_LINE_ZSH" "$ZSHRC" 2>/dev/null; then
        echo "    [*] Removing completion from $ZSHRC..."
        sed -i.bak "/^${COMMENT_LINE_ZSH}$/d; /^${COMPLETION_LINE_ZSH}$/d" "$ZSHRC" 2>/dev/null
        sed -i.bak -E '/^[[:space:]]*$/N;/\n[[:space:]]*$/D' "$ZSHRC" 2>/dev/null
        rm -f "${ZSHRC}.bak" 2>/dev/null
        echo "    [✔] Completion removed from $ZSHRC."
    else
        echo "    [i] Argus completion line not found in $ZSHRC. Skipping."
    fi
else
    echo "    [i] $ZSHRC not found. Skipping Zsh completion removal."
fi

# --- For Fish ---
FISH_COMPLETION_FILE="$HOME/.config/fish/completions/argus.fish"

if [[ -f "$FISH_COMPLETION_FILE" ]]; then
    echo "    [*] Removing Fish completion file: $FISH_COMPLETION_FILE..."
    rm -f "$FISH_COMPLETION_FILE"
    if [[ $? -eq 0 ]]; then
        echo "    [✔] Fish completion file removed."
    else
        echo "    [✖] Failed to remove Fish completion file. Check permissions."
    fi
else
    echo "    [i] Fish completion file not found ($FISH_COMPLETION_FILE). Skipping."
fi

echo "[✔] Argus uninstallation complete."
