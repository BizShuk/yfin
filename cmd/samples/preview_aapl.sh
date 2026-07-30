#!/usr/bin/env bash

# Current single-ticker CLI examples. These commands contact Yahoo Finance.
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd "$script_dir/../.." && pwd)"
yfin_cmd="${YFIN_CMD:-$repo_root/yfin}"
ticker="${TICKER:-AAPL}"

if [[ ! -x "$yfin_cmd" ]]; then
    echo "Missing executable: $yfin_cmd" >&2
    echo "Build it with: go build -o yfin ." >&2
    exit 1
fi

config_args=()
if [[ -n "${YFIN_CONFIG:-}" ]]; then
    config_args=(--config "$YFIN_CONFIG")
fi

run() {
    local title="$1"
    shift
    echo
    echo "== $title =="
    "$@"
}

run "Effective config" \
    "$yfin_cmd" "${config_args[@]}" config --print-effective

run "Connectivity check with response preview" \
    "$yfin_cmd" "${config_args[@]}" scrape \
    --check \
    --preview \
    --ticker "$ticker" \
    --endpoint key-statistics

run "Multi-endpoint JSON extraction" \
    "$yfin_cmd" "${config_args[@]}" scrape \
    --preview-json \
    --ticker "$ticker" \
    --endpoints key-statistics,financials,analysis,profile

run "News parser" \
    "$yfin_cmd" "${config_args[@]}" scrape \
    --preview-news \
    --ticker "$ticker"

run "Snapshot quote" \
    "$yfin_cmd" "${config_args[@]}" quote --tickers "$ticker"
