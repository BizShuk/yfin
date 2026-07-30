#!/usr/bin/env bash

# Current batch examples for daily bars and explicit scrape extraction.
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd "$script_dir/../.." && pwd)"
yfin_cmd="${YFIN_CMD:-$repo_root/yfin}"
start_date="${START_DATE:-2024-01-01}"
end_date="${END_DATE:-2024-01-08}"
output_dir="${OUTPUT_DIR:-$repo_root/tmp/sample-batch-output}"

if [[ ! -x "$yfin_cmd" ]]; then
    echo "Missing executable: $yfin_cmd" >&2
    echo "Build it with: go build -o yfin ." >&2
    exit 1
fi

config_args=()
if [[ -n "${YFIN_CONFIG:-}" ]]; then
    config_args=(--config "$YFIN_CONFIG")
fi

universe_file="$(mktemp)"
trap 'rm -f "$universe_file"' EXIT
printf '%s\n' AAPL MSFT GOOGL > "$universe_file"

echo "== Daily bars with local JSON export =="
"$yfin_cmd" "${config_args[@]}" \
    --concurrency 3 \
    --qps 2 \
    pull \
    --universe-file "$universe_file" \
    --start "$start_date" \
    --end "$end_date" \
    --adjusted split_dividend \
    --out json \
    --out-dir "$output_dir"

echo
echo "== Explicit scrape extraction per ticker =="
while IFS= read -r ticker; do
    "$yfin_cmd" "${config_args[@]}" \
        --qps 1 \
        scrape \
        --preview-json \
        --ticker "$ticker" \
        --endpoints key-statistics
done < "$universe_file"
