#!/usr/bin/env sh
set -eu

APP_NAME="moonawak3-minecraft"
DIST_DIR="dist"

mkdir -p "$DIST_DIR"

build() {
	GOOS="$1"
	GOARCH="$2"
	OUTPUT="$3"

	echo "Building $OUTPUT..."
	GOOS="$GOOS" GOARCH="$GOARCH" go build -trimpath -o "$DIST_DIR/$OUTPUT" .
}

build windows amd64 "$APP_NAME-windows-amd64.exe"
build darwin arm64 "$APP_NAME-macos-arm64"
build darwin amd64 "$APP_NAME-macos-amd64"
build linux amd64 "$APP_NAME-linux-amd64"

echo "Done. Binaries are in $DIST_DIR/"
