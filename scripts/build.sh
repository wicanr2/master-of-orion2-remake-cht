#!/usr/bin/env bash
# 在 docker 內編譯(依 CLAUDE.md:編譯一律 docker)。
# 用法:scripts/build.sh [go build 目標,預設 ./...]
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
IMAGE="golang:1.25-bookworm"

exec docker run --rm \
  -v "${REPO_ROOT}:/src" \
  -v "${REPO_ROOT}/.docker-cache/go:/go" \
  -w /src \
  -e CGO_ENABLED=0 \
  "${IMAGE}" \
  go build -buildvcs=false "${@:-./...}"
