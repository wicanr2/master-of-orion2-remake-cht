# Build/test 環境。Phase 1 為純資料層(無 ebiten/GL 依賴),用官方 golang 即可。
# 之後 Phase 2 引入 ebiten 時,再於此加 X11/ALSA/mesa 等圖形依賴。
FROM golang:1.25-bookworm

WORKDIR /src
ENV CGO_ENABLED=0 \
    GOFLAGS=-mod=mod
