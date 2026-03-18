#!/usr/bin/env bash
# Renders every .mmd file under assets/diagrams/ to an SVG of the same name.
# Requires @mermaid-js/mermaid-cli (mmdc), available via npx.
#
# Usage:
#   .github/scripts/render-diagrams.sh

set -euo pipefail

DIAGRAMS_DIR="$(git rev-parse --show-toplevel)/assets/diagrams"

find "$DIAGRAMS_DIR" -name '*.mmd' | while read -r src; do
  out="${src%.mmd}.svg"
  echo "Rendering $src → $out"
  npx --yes @mermaid-js/mermaid-cli -i "$src" -o "$out"
done
