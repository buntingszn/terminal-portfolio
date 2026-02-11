#!/usr/bin/env bash
# tui-stats.sh — Parse TUI analytics JSONL and display visitor stats.
#
# Requires: jq
#
# Usage:
#   ./scripts/tui-stats.sh [options]
#
# Options:
#   -f FILE    Analytics file (default: /opt/terminal-portfolio/analytics.jsonl)
#   -d DAYS    Only show last N days (default: all)
#   --today    Shortcut for -d 1
#   --week     Shortcut for -d 7
#   --json     Output raw JSON instead of formatted table

set -euo pipefail

FILE="/opt/terminal-portfolio/analytics.jsonl"
DAYS=""
JSON_OUTPUT=false

while [[ $# -gt 0 ]]; do
    case "$1" in
        -f)       FILE="$2"; shift 2 ;;
        -d)       DAYS="$2"; shift 2 ;;
        --today)  DAYS=1; shift ;;
        --week)   DAYS=7; shift ;;
        --json)   JSON_OUTPUT=true; shift ;;
        -h|--help)
            echo "Usage: $0 [options]"
            echo "  -f FILE    Analytics file (default: /opt/terminal-portfolio/analytics.jsonl)"
            echo "  -d DAYS    Only show last N days (default: all)"
            echo "  --today    Shortcut for -d 1"
            echo "  --week     Shortcut for -d 7"
            echo "  --json     Output raw JSON instead of formatted table"
            exit 0
            ;;
        *)
            echo "Unknown option: $1" >&2
            exit 1
            ;;
    esac
done

if ! command -v jq &>/dev/null; then
    echo "Error: jq is required but not installed." >&2
    exit 1
fi

if [[ ! -f "$FILE" ]]; then
    echo "Error: Analytics file not found: $FILE" >&2
    exit 1
fi

# Build a date filter for jq if DAYS is set.
DATE_FILTER="."
PERIOD_LABEL="all time"
if [[ -n "$DAYS" ]]; then
    CUTOFF=$(date -u -d "$DAYS days ago" +%Y-%m-%dT%H:%M:%S 2>/dev/null || \
             date -u -v-"${DAYS}"d +%Y-%m-%dT%H:%M:%S 2>/dev/null)
    DATE_FILTER="select(.ts >= \"$CUTOFF\")"
    if [[ "$DAYS" == "1" ]]; then
        PERIOD_LABEL="today"
    else
        PERIOD_LABEL="last $DAYS days"
    fi
fi

# Slurp all events matching the date filter.
EVENTS=$(jq -c "[$DATE_FILTER]" "$FILE")

if $JSON_OUTPUT; then
    jq -n --argjson events "$EVENTS" '{
        summary: {
            sessions: ([$events[] | select(.type == "session_start")] | length),
            unique_ips: ([$events[] | select(.type == "session_start") | .ip] | unique | length),
            avg_duration_ms: (
                [$events[] | select(.type == "session_end") | .duration_ms] |
                if length > 0 then (add / length | round) else 0 end
            )
        },
        section_views: (
            [$events[] | select(.type == "section_view")] | group_by(.section) |
            map({
                section: .[0].section,
                views: length,
                avg_duration_ms: ([.[].duration_ms] | add / length | round)
            }) | sort_by(-.views)
        ),
        sessions_by_day: (
            [$events[] | select(.type == "session_start")] |
            group_by(.ts[:10]) |
            map({date: .[0].ts[:10], count: length}) |
            sort_by(.date) | reverse
        ),
        top_ips: (
            [$events[] | select(.type == "session_start")] |
            group_by(.ip) |
            map({ip: .[0].ip, sessions: length}) |
            sort_by(-.sessions) | .[0:10]
        )
    }'
    exit 0
fi

# --- Formatted output ---

echo "TUI Analytics \u2014 $FILE"
printf '\u2550%.0s' {1..40}
echo ""
echo ""

# Summary
SESSIONS=$(echo "$EVENTS" | jq '[.[] | select(.type == "session_start")] | length')
UNIQUE_IPS=$(echo "$EVENTS" | jq '[.[] | select(.type == "session_start") | .ip] | unique | length')
AVG_DURATION_MS=$(echo "$EVENTS" | jq '[.[] | select(.type == "session_end") | .duration_ms] | if length > 0 then (add / length | round) else 0 end')

# Convert ms to human-readable duration.
format_duration() {
    local ms=$1
    local secs=$((ms / 1000))
    local mins=$((secs / 60))
    secs=$((secs % 60))
    if [[ $mins -gt 0 ]]; then
        echo "${mins}m${secs}s"
    else
        echo "${secs}s"
    fi
}

echo "Summary ($PERIOD_LABEL)"
printf "  Sessions:     %6d\n" "$SESSIONS"
printf "  Unique IPs:   %6d\n" "$UNIQUE_IPS"
printf "  Avg duration: %6s\n" "$(format_duration "$AVG_DURATION_MS")"
echo ""

# Section Views
echo "Section Views"
echo "$EVENTS" | jq -r '
    [.[] | select(.type == "section_view")] | group_by(.section) |
    map({
        section: .[0].section,
        views: length,
        avg_ms: ([.[].duration_ms] | add / length | round)
    }) | sort_by(-.views)[] |
    "  \(.section)\t\(.views)\t\(.avg_ms)"
' | while IFS=$'\t' read -r section views avg_ms; do
    avg_human=$(format_duration "$avg_ms")
    printf "  %-10s %4d  (avg %s)\n" "$section" "$views" "$avg_human"
done
echo ""

# Sessions by Day (last 14 entries)
echo "Sessions by Day"
echo "$EVENTS" | jq -r '
    [.[] | select(.type == "session_start")] |
    group_by(.ts[:10]) |
    map({date: .[0].ts[:10], count: length}) |
    sort_by(.date) | reverse | .[0:14][] |
    "  \(.date)   \(.count)"
'
echo ""

# Top IPs
echo "Top IPs"
echo "$EVENTS" | jq -r '
    [.[] | select(.type == "session_start")] |
    group_by(.ip) |
    map({ip: .[0].ip, sessions: length}) |
    sort_by(-.sessions) | .[0:10][] |
    "  \(.ip)\t\(.sessions)"
' | while IFS=$'\t' read -r ip sessions; do
    printf "  %-20s %4d\n" "$ip" "$sessions"
done
