#!/bin/sh
set -e
REPO="andrearcaina/pathfinder"
BIN="pathfinder"

# get the latest tag from GitHub
TAG=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

# detect OS/Arch
OS="$(uname -s)"
ARCH="$(uname -m)"
EXT="tar.gz"

# detects architecture for naming consistency
case "$OS" in
    Linux*)  OS="Linux" ;;
    Darwin*) OS="Darwin" ;;
    MINGW*|MSYS*) OS="Windows"; EXT="zip"; BIN="${BIN}.exe" ;;
    *) echo "Unsupported OS: $OS"; exit 1 ;;
esac

# download and extract
URL="https://github.com/$REPO/releases/download/$TAG/${BIN%.*}_${OS}_${ARCH}.$EXT"
echo "Downloading $TAG for $OS..."

if [ "$EXT" = "zip" ]; then
    # windows: download zip, unzip, install to ~/bin (no sudo needed)
    curl -sL "$URL" -o tmp.zip
    unzip -qo tmp.zip "$BIN" && rm tmp.zip
    mkdir -p ~/bin && mv "$BIN" ~/bin/
    echo "Installed to ~/bin/$BIN (Ensure ~/bin is in your PATH)"
else
    # mac/linux: stream tar, sudo move to /usr/local/bin
    curl -sL "$URL" | tar xz
    sudo mv "$BIN" /usr/local/bin/
    echo "Installed to /usr/local/bin/$BIN"
fi
