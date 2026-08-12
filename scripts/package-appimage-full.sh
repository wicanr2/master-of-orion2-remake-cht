#!/usr/bin/env bash
# 完整本機測試版 AppImage:自帶【遊戲資料子集 + i18n 譯表 + CJK 字型 + 自訂 AppRun】,
# 啟動即進中文 -game,免下 -data。
#
# ⚠ 產出含版權遊戲資料,僅供【本機自用測試】,dist-all/ 已 gitignore,絕不入 repo/散布。
#   版權隔離:committed 的 package-appimage.sh 維持不含資料;本檔為本機 full build。
#
# 用法: MOO2_DATA=<遊戲資料夾> MOO2_FONT=<CJK字型.ttc> scripts/package-appimage-full.sh
#   預設 MOO2_DATA=/home/anr2/moo2-private-build/gamedata/mastori2
#        MOO2_FONT=/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc(OFL,可再散布)
# 產出: dist-all/MasterOfOrion2-cht-full-x86_64.AppImage
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
IMAGE="moo2-ebiten"
DIST_DIR="${MOO2_DIST_DIR:-${REPO_ROOT}/dist-all}"
TOOLS_CACHE="${REPO_ROOT}/.docker-cache/appimage-tools"
APP_NAME="MasterOfOrion2-cht-full"
DATA_DIR="${MOO2_DATA:-/home/anr2/moo2-private-build/gamedata/mastori2}"
FONT_FILE="${MOO2_FONT:-/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc}"

[ -d "${DATA_DIR}" ] || { echo "找不到遊戲資料夾: ${DATA_DIR}"; exit 1; }
[ -f "${FONT_FILE}" ] || { echo "找不到字型: ${FONT_FILE}"; exit 1; }

# 完整包帶 55 個**正常玩家路徑**會讀取的 LBX：由 cmd/moo2 靜態消費端與封裝後
# 35 張畫廊交叉確認。它不是把 373 個原版檔全塞入，而是不允許 INPUT／MULTIGM／
# COMBAT 等正常畫面悄悄降級成後備版。`stardb.lbx` 只在 resolver 測試作假檔，
# 正版資料不存在，故刻意不列入。
LBX_LIST="amebafin anatkfin antaroom anwinfin bldg0 bldg1 bldg2 bldg3 bldg4 bldg5 buffer0 cmbtsfx cmbtshp colbldg colgcbt colony colony2 colpups colroads colsum colveggi combat confirm council design dimtvfin diplomat fleet game genwinfn help herodata inbox info intro loserfin mainmenu multigm newgame officer orionfin planets plntdfin plntsum raceopt races racesel science sound starbg stream streamhd techsel turnsum wininfin"

for required_dir in "${DIST_DIR}" "${TOOLS_CACHE}" "${REPO_ROOT}/.docker-cache/go"; do
  if [[ ! -d "${required_dir}" ]]; then
    echo "缺少既有目錄: ${required_dir}；拒絕在主機端自行建立。" >&2
    exit 1
  fi
done

if ! docker image inspect "${IMAGE}" >/dev/null 2>&1; then
  echo "找不到既有打包映像: ${IMAGE}" >&2
  echo "請先依 docker/Dockerfile.ebiten 準備可重現工具鏈；本腳本不會自行下載或另建映像。" >&2
  exit 1
fi
for tool in linuxdeploy-x86_64.AppImage appimagetool-x86_64.AppImage runtime-x86_64; do
  if [[ ! -x "${TOOLS_CACHE}/${tool}" ]]; then
    echo "缺少離線 AppImage 工具快取: ${TOOLS_CACHE}/${tool}" >&2
    echo "請以受控、明確授權的下載步驟補齊快取後重跑。" >&2
    exit 1
  fi
done

docker run --rm --network none --memory 4g --cpus 2 --pids-limit 256 \
  -u "$(id -u):$(id -g)" \
  -e GOPATH=/go -e GOMODCACHE=/go/pkg/mod -e GOCACHE=/go/build-cache \
  -v "${REPO_ROOT}:/src:ro" \
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
    export PATH=/usr/local/go/bin:$PATH
    APPDIR=/tmp/AppDir
    RES="${APPDIR}/usr/share/moo2"

    echo "== [1/6] go build cmd/moo2 + cmd/moo2sim (CGO_ENABLED=${CGO_ENABLED}) =="
    mkdir -p "${APPDIR}/usr/bin" "${RES}"
    go build -buildvcs=false -ldflags="-s -w" -o "${APPDIR}/usr/bin/moo2" ./cmd/moo2
    go build -buildvcs=false -ldflags="-s -w" -o "${APPDIR}/usr/bin/moo2sim" ./cmd/moo2sim

    echo "== [2/6] 打包 i18n 譯表 + 字型 =="
    mkdir -p "${RES}/assets"
    cp -r assets/i18n "${RES}/assets/i18n"
    cp -r assets/fonts "${RES}/assets/fonts" 2>/dev/null || true
    cp /font.ttc "${RES}/font.ttc"

    echo "== [3/6] 打包遊戲資料子集(僅 -game 需要的 LBX,大小寫不敏感)=="
    mkdir -p "${RES}/gamedata"
    expected_lbx=0
    for name in ${LBX_LIST}; do
      expected_lbx=$((expected_lbx + 1))
      # 來源可能大小寫不一,逐一比對複製。
      found=""
      for cand in /gamedata/${name}.lbx /gamedata/${name}.LBX; do
        [ -f "$cand" ] && found="$cand" && break
      done
      if [ -z "$found" ]; then
        # 大小寫不敏感搜尋。
        found="$(find /gamedata -maxdepth 1 -type f -iname "${name}.lbx" -print -quit)"
      fi
      [ -n "$found" ] || { echo "缺少完整版必要資料: ${name}.lbx" >&2; exit 1; }
      cp "$found" "${RES}/gamedata/$(basename "$found" | tr "a-z" "A-Z")"
    done
    actual_lbx="$(find "${RES}/gamedata" -maxdepth 1 -type f -iname "*.lbx" | wc -l)"
    [ "$actual_lbx" = "$expected_lbx" ] || {
      echo "完整版 LBX 數量不符: 預期 $expected_lbx，實得 $actual_lbx" >&2
      exit 1
    }
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
    RT=/tools/runtime-x86_64
    [ -x "$LD" ] && [ -x "$AT" ] && [ -x "$RT" ]
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
    OUT="/dist/.${APP_NAME}-x86_64.AppImage.tmp"
    rm -f "$OUT"
    "$AT" --appimage-extract-and-run --runtime-file "$RT" "${APPDIR}" "$OUT"
    mv "$OUT" "/dist/${APP_NAME}-x86_64.AppImage"
  '

echo "產出: ${DIST_DIR}/${APP_NAME}-x86_64.AppImage"
ls -la "${DIST_DIR}/${APP_NAME}-x86_64.AppImage"
