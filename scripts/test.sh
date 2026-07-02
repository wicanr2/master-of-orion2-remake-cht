#!/usr/bin/env bash
# 在 docker 內跑 go test(依 CLAUDE.md:編譯/測試一律 docker)。
# 用法:scripts/test.sh [額外的 go test 參數]
#   MOO2_LBX_TEST=/abs/path/to/FILE.LBX scripts/test.sh   # 一併跑真實 .lbx 測試
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
IMAGE="golang:1.25-bookworm"

# 若指定了真實 .lbx,把它掛進容器並轉成容器內路徑。
MOUNT_ARGS=()
ENV_ARGS=()
if [[ -n "${MOO2_LBX_TEST:-}" ]]; then
  MOUNT_ARGS+=(-v "${MOO2_LBX_TEST}:/testdata/real.lbx:ro")
  ENV_ARGS+=(-e "MOO2_LBX_TEST=/testdata/real.lbx")
fi

exec docker run --rm \
  -v "${REPO_ROOT}:/src" \
  -v "${REPO_ROOT}/.docker-cache/go:/go" \
  -w /src \
  -e CGO_ENABLED=0 \
  "${ENV_ARGS[@]}" \
  "${MOUNT_ARGS[@]}" \
  "${IMAGE}" \
  go test "${@:-./...}"
