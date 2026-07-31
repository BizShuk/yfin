#!/usr/bin/env bash
set -euo pipefail

yfin_live_integration=false
yfin_race=false
yfin_verbose=false

while (($# > 0)); do
    case "$1" in
        --live-integration)
            yfin_live_integration=true
            ;;
        --race)
            yfin_race=true
            ;;
        --verbose)
            yfin_verbose=true
            ;;
        --help)
            printf 'Usage: %s [--live-integration] [--race] [--verbose]\n' "$0"
            exit 0
            ;;
        *)
            printf 'Unknown option: %s\n' "$1" >&2
            exit 2
            ;;
    esac
    shift
done

yfin_test_args=(-count=1)
if [[ "${yfin_race}" == true ]]; then
    yfin_test_args+=(-race)
fi
if [[ "${yfin_verbose}" == true ]]; then
    yfin_test_args+=(-v)
fi

printf 'Running deterministic yfin tests...\n'
go test "${yfin_test_args[@]}" ./...

printf 'Compiling integration-tagged packages without live requests...\n'
go test "${yfin_test_args[@]}" -tags=integration ./... -run '^$'

if [[ "${yfin_live_integration}" == true ]]; then
    printf 'Running live integration tests against external providers...\n'
    go test "${yfin_test_args[@]}" -tags=integration ./...
fi

printf 'All requested yfin tests passed.\n'
