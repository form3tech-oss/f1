#!/usr/bin/env bash

set -euo pipefail

version="$1"
readonly version

if [[ -x tools/golangci-lint && $(tools/golangci-lint version) == *"$version"* ]]; then
  echo "$version already installed"
  exit 0
fi

os=$(uname -s | tr '[:upper:]' '[:lower:]')
readonly os

arch=$(go env GOARCH)
readonly arch

asset="golangci-lint-$version-$os-$arch"
readonly asset

url="https://github.com/golangci/golangci-lint/releases/download/v${version}/$asset.tar.gz"
readonly url

tarball="$(mktemp)"
readonly tarball

mkdir -p tools/
curl -sSfL "$url" -o "$tarball"

signatures="$(dirname "$0")/golangci-lint-$version-checksums.txt"
readonly signatures
expected_signature=$(grep "$asset" "$signatures" | cut -d ' ' -f 1)
readonly expected_signature

echo "$expected_signature $tarball" | sha256sum --check || { echo "Checksum verification failed for $asset"; exit 1; }

tar xf "$tarball" -C tools --strip-components 1 "$asset/golangci-lint"
rm -rf "$tarball"
tools/golangci-lint version
