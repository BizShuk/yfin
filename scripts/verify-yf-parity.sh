#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${repo_root}"

ticker="${1:-AAPL}"
day="$(date -u +%F)"
python_bin="${PYTHON_BIN:-${HOME}/.venv/bin/python}"
if [[ ! -x "${python_bin}" ]]; then
  python_bin="python3"
fi

"${python_bin}" skills/scripts/all_ticker_yf.py \
  --ticker "${ticker}" \
  --force \
  --max-workers 1
go run . batch \
  --ticker "${ticker}" \
  --force \
  --max-workers 1
"${python_bin}" skills/scripts/compare_yfin_parity.py \
  --python-root "${HOME}/.config/stock/data/raw" \
  --go-root "${HOME}/.config/yfin/data/raw" \
  --ticker "${ticker}" \
  --day "${day}"
