#!/usr/bin/env sh
set -eu

APP_NAME="moonawak3-minecraft"
DIST_DIR="dist"
VERSION="${VERSION:-v0.0.1}"
LDFLAGS="-X moonawak3-minecraft/internal/version.Current=$VERSION"

rm -rf "$DIST_DIR"
mkdir -p "$DIST_DIR"

build() {
	GOOS="$1"
	GOARCH="$2"
	OUTPUT="$3"

	echo "Building $OUTPUT ($VERSION)..."
	GOOS="$GOOS" GOARCH="$GOARCH" go build -trimpath -ldflags "$LDFLAGS" -o "$DIST_DIR/$OUTPUT" .
}

build windows amd64 "$APP_NAME-windows-amd64.exe"
build darwin arm64 "$APP_NAME-macos-arm64"
build darwin amd64 "$APP_NAME-macos-amd64"
build linux amd64 "$APP_NAME-linux-amd64"

(cd "$DIST_DIR" && shasum -a 256 * > checksums.txt)

echo "Done. Binaries are in $DIST_DIR/"
