#!/bin/sh
# cosmobar installer — downloads the latest release binary into ~/.local/bin.
set -eu

REPO="oleksbard/cosmobar"
BIN_DIR="${COSMOBAR_BIN_DIR:-$HOME/.local/bin}"

os="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "$os" in
  darwin) os="darwin" ;;
  linux)  os="linux" ;;
  *) echo "cosmobar: unsupported OS: $os" >&2; exit 1 ;;
esac

arch="$(uname -m)"
case "$arch" in
  x86_64|amd64) arch="amd64" ;;
  arm64|aarch64) arch="arm64" ;;
  *) echo "cosmobar: unsupported arch: $arch" >&2; exit 1 ;;
esac

asset="cosmobar_${os}_${arch}.tar.gz"
base="https://github.com/${REPO}/releases/latest/download"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

echo "cosmobar: downloading $asset ..."
curl -fsSL "$base/$asset" -o "$tmp/$asset"
curl -fsSL "$base/checksums.txt" -o "$tmp/checksums.txt"

echo "cosmobar: verifying checksum ..."
expected="$(grep " $asset\$" "$tmp/checksums.txt" | awk '{print $1}')"
if command -v sha256sum >/dev/null 2>&1; then
  actual="$(sha256sum "$tmp/$asset" | awk '{print $1}')"
else
  actual="$(shasum -a 256 "$tmp/$asset" | awk '{print $1}')"
fi
if [ "$expected" != "$actual" ]; then
  echo "cosmobar: checksum mismatch; aborting." >&2
  exit 1
fi

tar -xzf "$tmp/$asset" -C "$tmp"
mkdir -p "$BIN_DIR"
install -m 0755 "$tmp/cosmobar" "$BIN_DIR/cosmobar"

echo "cosmobar: installed to $BIN_DIR/cosmobar"
case ":$PATH:" in
  *":$BIN_DIR:"*) ;;
  *) echo "cosmobar: add $BIN_DIR to your PATH, then run: cosmobar init" ;;
esac
echo "cosmobar: run 'cosmobar init' to wire it into Claude Code."
