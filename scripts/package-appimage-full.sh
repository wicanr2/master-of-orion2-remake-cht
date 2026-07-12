#!/usr/bin/env bash
# 完整本機測試版 AppImage:自帶【遊戲資料子集 + i18n 譯表 + CJK 字型 + 自訂 AppRun】,
# 啟動即進中文 -game,免下 -data。
#
# ⚠ 產出含版權遊戲資料,僅供【本機自用測試】,dist/ 已 gitignore,絕不入 repo/散布。
#   版權隔離:committed 的 package-appimage.sh 維持不含資料;本檔為本機 full build。
#
# 用法: MOO2_DATA=<遊戲資料夾> MOO2_FONT=<CJK字型.ttc> scripts/package-appimage-full.sh
#   預設 MOO2_DATA=/home/anr2/moo2-private-build/gamedata/mastori2
#        MOO2_FONT=/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc(OFL,可再散布)
# 產出: dist/MasterOfOrion2-cht-full-x86_64.AppImage
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
IMAGE="moo2-ebiten"
DIST_DIR="${REPO_ROOT}/dist"
TOOLS_CACHE="${REPO_ROOT}/.docker-cache/appimage-tools"
APP_NAME="MasterOfOrion2-cht-full"
DATA_DIR="${MOO2_DATA:-/home/anr2/moo2-private-build/gamedata/mastori2}"
FONT_FILE="${MOO2_FONT:-/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc}"

[ -d "${DATA_DIR}" ] || { echo "找不到遊戲資料夾: ${DATA_DIR}"; exit 1; }
[ -f "${FONT_FILE}" ] || { echo "找不到字型: ${FONT_FILE}"; exit 1; }

# -game 實際載入的 LBX(取自 cmd/moo2 原始碼 grep;只打包需要的,避免 324M 全帶)。
# 含音樂/音效(stream/streamhd/sound)與英雄資料(herodata)——本機 full build 不省體積,
# 缺這幾個會導致「沒音樂 / 傭兵池空無人可雇」(2026-07-12 實測 issue #1/#4)。
LBX_LIST="buffer0 colsum council design diplomat fleet help info mainmenu newgame officer plntsum raceopt races science stardb techsel turnsum game stream streamhd sound herodata"

mkdir -p "${DIST_DIR}" "${TOOLS_CACHE}" "${REPO_ROOT}/.docker-cache/go"

if ! docker image inspect "${IMAGE}" >/dev/null 2>&1; then
  docker build -t "${IMAGE}" -f "${REPO_ROOT}/docker/Dockerfile.ebiten" "${REPO_ROOT}"
fi

docker run --rm \
  -v "${REPO_ROOT}:/src" \
  -v "${REPO_ROOT}/.docker-cache/go:/go" \
  -v "${TOOLS_CACHE}:/tools" \
  -v "${DIST_DIR}:/dist" \
  -v "${DATA_DIR}:/gamedata:ro" \
  -v "${FONT_FILE}:/font.ttc:ro" \
  -w /src \
  -e "APP_NAME=${APP_NAME}" \
  -e "LBX_LIST=${LBX_LIST}" \
  "${IMAGE}" \
  bash -eu -o pipefail -c '
    APPDIR=/tmp/AppDir
    RES="${APPDIR}/usr/share/moo2"

    echo "== [1/6] go build cmd/moo2 (CGO_ENABLED=${CGO_ENABLED}) =="
    mkdir -p "${APPDIR}/usr/bin" "${RES}"
    go build -buildvcs=false -ldflags="-s -w" -o "${APPDIR}/usr/bin/moo2" ./cmd/moo2

    echo "== [2/6] 打包 i18n 譯表 + 字型 =="
    mkdir -p "${RES}/assets"
    cp -r assets/i18n "${RES}/assets/i18n"
    cp -r assets/fonts "${RES}/assets/fonts" 2>/dev/null || true
    cp /font.ttc "${RES}/font.ttc"

    echo "== [3/6] 打包遊戲資料子集(僅 -game 需要的 LBX,大小寫不敏感)=="
    mkdir -p "${RES}/gamedata"
    for name in ${LBX_LIST}; do
      # 來源可能大小寫不一,逐一比對複製。
      found=""
      for cand in /gamedata/${name}.lbx /gamedata/${name}.LBX; do
        [ -f "$cand" ] && found="$cand" && break
      done
      if [ -z "$found" ]; then
        # 大小寫不敏感搜尋。
        found="$(find /gamedata -maxdepth 1 -iname "${name}.lbx" | head -1 || true)"
      fi
      if [ -n "$found" ]; then
        cp "$found" "${RES}/gamedata/$(basename "$found" | tr "a-z" "A-Z")"
      else
        echo "   (略過缺檔: ${name}.lbx)"
      fi
    done
    echo "   遊戲資料合計: $(du -sh "${RES}/gamedata" | cut -f1)"

    echo "== [4/6] .desktop + 佔位圖示(不含版權美術)=="
    mkdir -p "${APPDIR}/usr/share/applications" "${APPDIR}/usr/share/icons/hicolor/256x256/apps"
    go run scripts/gen-icon.go "${APPDIR}/usr/share/icons/hicolor/256x256/apps/moo2.png"
    cat > "${APPDIR}/usr/share/applications/moo2.desktop" <<EOF
[Desktop Entry]
Type=Application
Name=Master of Orion 2 (中文化 Remake)
Comment=銀河霸主2 go/ebiten remake(完整測試版,已內含資料)
Exec=moo2
Icon=moo2
Categories=Game;StrategyGame;
Terminal=false
EOF

    echo "== [5/6] linuxdeploy 掃依賴(libGL/libX11 等)=="
    LD=/tools/linuxdeploy-x86_64.AppImage
    AT=/tools/appimagetool-x86_64.AppImage
    [ -x "$LD" ] || { curl -sSL -o "$LD" https://github.com/linuxdeploy/linuxdeploy/releases/download/continuous/linuxdeploy-x86_64.AppImage; chmod +x "$LD"; }
    [ -x "$AT" ] || { curl -sSL -o "$AT" https://github.com/AppImage/appimagetool/releases/download/continuous/appimagetool-x86_64.AppImage; chmod +x "$AT"; }
    cd /tmp
    "$LD" --appimage-extract-and-run \
      --appdir "${APPDIR}" \
      --executable "${APPDIR}/usr/bin/moo2" \
      --desktop-file "${APPDIR}/usr/share/applications/moo2.desktop" \
      --icon-file "${APPDIR}/usr/share/icons/hicolor/256x256/apps/moo2.png"

    echo "== 覆寫 AppRun:啟動即中文 -game,自帶資料/字型/譯表 =="
    rm -f "${APPDIR}/AppRun"
    cat > "${APPDIR}/AppRun" <<"EOF"
#!/bin/bash
HERE="$(dirname "$(readlink -f "$0")")"
RES="$HERE/usr/share/moo2"
cd "$RES"   # 讓 assets/i18n 相對路徑可解析
exec "$HERE/usr/bin/moo2" -game -lang zh -data "$RES/gamedata" -font "$RES/font.ttc" "$@"
EOF
    chmod +x "${APPDIR}/AppRun"

    echo "== [6/6] appimagetool 打包 =="
    "$AT" --appimage-extract-and-run "${APPDIR}" "/dist/${APP_NAME}-x86_64.AppImage"
  '

echo "產出: ${DIST_DIR}/${APP_NAME}-x86_64.AppImage"
ls -la "${DIST_DIR}/${APP_NAME}-x86_64.AppImage"
