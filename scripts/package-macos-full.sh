#!/usr/bin/env bash
# 本機完整 macOS universal 測試包；資料與音訊均為私有輸入，不可公開散布。
set -euo pipefail

moo2_script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MOO2_FULL=1 exec "${moo2_script_dir}/package-macos.sh" "$@"
