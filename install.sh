#!/bin/sh
REPO="andrearcaina/pathfinder"
BINARY="pathfinder"

# get the latest tag from GitHub
TAG=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

# detect OS/Arch
OS=$(uname -s) # linux or darwin
ARCH=$(uname -m) # x86_64 or arm64

# download and Unzip
URL="https://github.com/$REPO/releases/download/$TAG/${BINARY}_${OS}_${ARCH}.tar.gz"
echo "Downloading $BINARY $TAG for $OS ($ARCH)..."

curl -L "$URL" | tar xz

# install
echo "Installing to /usr/local/bin (password may be required)..."
sudo mv $BINARY /usr/local/bin/
echo "Done! Run '$BINARY version' to test."
