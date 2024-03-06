#!/usr/bin/env bash

set -euo pipefail

version="$1"
readonly version

os=$(uname -s | tr '[:upper:]' '[:lower:]')
readonly os

arch=$(go env GOARCH) # note that uname -m gives unexpected results in an f3 shell

asset="golangci-lint-$version-$os-$arch"
readonly asset

url="https://github.com/golangci/golangci-lint/releases/download/v${version}/$asset.tar.gz"
readonly url

tarball="$(mktemp)"
readonly tarball

if [[ -x tools/golangci-lint && $(tools/golangci-lint version) == *"$version"* ]]; then
  echo "$version already installed"
  exit 0
fi

mkdir -p tools/
curl -sSfL "$url" -o "$tarball"
tar xf "$tarball" -C tools --strip-components 1 "$asset/golangci-lint"
rm -rf "$tarball"
tools/golangci-lint version
