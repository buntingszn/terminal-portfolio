#!/usr/bin/env bash
set -euo pipefail

# ──────────────────────────────────────────────
# validate-data.sh — Validate all JSON data files
# ──────────────────────────────────────────────

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
DATA_DIR="$PROJECT_ROOT/data/content"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BOLD='\033[1m'
RESET='\033[0m'

# Counters
PASS=0
FAIL=0
TOTAL=0

pass() {
  echo -e "  ${GREEN}✓${RESET} $1"
  PASS=$((PASS + 1))
  TOTAL=$((TOTAL + 1))
}

fail() {
  echo -e "  ${RED}✗${RESET} $1"
  FAIL=$((FAIL + 1))
  TOTAL=$((TOTAL + 1))
}

# ── Check that jq is available ───────────────────────────────
if ! command -v jq &>/dev/null; then
  echo -e "${RED}Error: jq is required but not installed.${RESET}"
  exit 1
fi

# ── Define files and their required top-level fields ─────────
declare -A REQUIRED_FIELDS
REQUIRED_FIELDS=(
  ["meta.json"]="version name title oneLiner siteUrl sshAddress sourceRepo"
  ["about.json"]="bio location status education"
  ["work.json"]="projects"
  ["cv.json"]="contact summary experience skills education"
  ["links.json"]="links"
)

FILES=("meta.json" "about.json" "work.json" "cv.json" "links.json")

echo -e "${BOLD}Validating data files...${RESET}"
echo ""

# ── Phase 1: Check file existence ────────────────────────────
echo -e "${BOLD}File existence${RESET}"
MISSING_FILES=()
for file in "${FILES[@]}"; do
  filepath="$DATA_DIR/$file"
  if [ -f "$filepath" ]; then
    pass "$file exists"
  else
    fail "$file is missing"
    MISSING_FILES+=("$file")
  fi
done

echo ""

# ── Phase 2: Validate JSON syntax ───────────────────────────
echo -e "${BOLD}JSON syntax${RESET}"
INVALID_JSON=()
for file in "${FILES[@]}"; do
  filepath="$DATA_DIR/$file"
  # Skip files that don't exist (already reported above)
  if [ ! -f "$filepath" ]; then
    fail "$file — skipped (file missing)"
    INVALID_JSON+=("$file")
    continue
  fi
  if jq empty "$filepath" 2>/dev/null; then
    pass "$file is valid JSON"
  else
    fail "$file has invalid JSON syntax"
    INVALID_JSON+=("$file")
  fi
done

echo ""

# ── Phase 3: Check required fields ──────────────────────────
echo -e "${BOLD}Required fields${RESET}"
for file in "${FILES[@]}"; do
  filepath="$DATA_DIR/$file"

  # Skip files that are missing or have invalid JSON
  if [[ " ${MISSING_FILES[*]:-} " == *" $file "* ]] || [[ " ${INVALID_JSON[*]:-} " == *" $file "* ]]; then
    fail "$file — skipped (missing or invalid)"
    continue
  fi

  fields=${REQUIRED_FIELDS[$file]}
  file_ok=true
  for field in $fields; do
    if jq -e "has(\"$field\")" "$filepath" &>/dev/null; then
      pass "$file has field \"$field\""
    else
      fail "$file is missing required field \"$field\""
      file_ok=false
    fi
  done
done

# ── Summary ──────────────────────────────────────────────────
echo ""
echo -e "${BOLD}────────────────────────────────${RESET}"
if [ "$FAIL" -eq 0 ]; then
  echo -e "${GREEN}${BOLD}All $TOTAL checks passed.${RESET}"
  exit 0
else
  echo -e "${RED}${BOLD}$FAIL of $TOTAL checks failed.${RESET}"
  exit 1
fi
