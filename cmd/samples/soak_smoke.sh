#!/usr/bin/env bash

# Short standalone soak run. This command contacts Yahoo Finance.
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd "$script_dir/../.." && pwd)"
universe_file="${UNIVERSE_FILE:-$repo_root/tests/testdata/universe/soak.txt}"

config_args=()
if [[ -n "${YFIN_CONFIG:-}" ]]; then
    config_args=(--config "$YFIN_CONFIG")
fi

cd "$repo_root"
go run ./cmd/soak \
    "${config_args[@]}" \
    --universe-file "$universe_file" \
    --endpoints key-statistics,news \
    --fallback auto \
    --duration "${SOAK_DURATION:-30s}" \
    --concurrency "${SOAK_CONCURRENCY:-2}" \
    --qps "${SOAK_QPS:-1}" \
    --probe-interval "${SOAK_PROBE_INTERVAL:-15m}" \
    --failure-rate "${SOAK_FAILURE_RATE:-0}"
