#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WEB_DIR="$(dirname "$SCRIPT_DIR")"
SERVE_DIR="${SERVE_DIR:-/var/www/terminal-portfolio}"

cd "$WEB_DIR"
npm run build

echo "Copying dist/ to $SERVE_DIR..."
mkdir -p "$SERVE_DIR"
rsync -a --delete dist/ "$SERVE_DIR/"
echo "Deploy complete."
