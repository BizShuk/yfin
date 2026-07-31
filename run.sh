#!/usr/bin/env bash
set -euo pipefail

config_root="${XDG_CONFIG_HOME:-${HOME}/.config}/yfin"

mkdir -p "${config_root}/data" "${config_root}/logs"

printf 'yfin metadata ready: %s\n' "${config_root}"
