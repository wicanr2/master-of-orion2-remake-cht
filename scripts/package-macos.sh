#!/usr/bin/env bash
# macOS universal 本機打包：以既有 u5cht/osxcross 產出未簽署的 .app + tar.gz。
#
# 公開版：scripts/package-macos.sh
# 本機完整版：scripts/package-macos-full.sh（透過 MOO2_FULL=1 呼叫本檔）
#
# 預設產物在 dist-all/。公開包不含原版資料／音訊／字型；完整版只供有相應授權的
# 本機測試，絕不可公開散布。這條 osxcross 路徑不取代真 macOS 的 codesign、Gatekeeper
# 或實機驗收。
set -euo pipefail

moo2_repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
moo2_image="u5cht/osxcross:latest"
moo2_dist_dir="${MOO2_DIST_DIR:-${moo2_repo_root}/dist-all}"
moo2_go_cache="${moo2_repo_root}/.docker-cache/go"
moo2_full_build="${MOO2_FULL:-0}"
moo2_app_name="MasterOfOrion2-cht"
moo2_data_dir="${MOO2_DATA:-/home/anr2/moo2-private-build/gamedata/mastori2}"
moo2_font_file="${MOO2_FONT:-/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc}"

# 與 Linux／Windows 完整包同源：55 個由靜態消費端與封裝畫廊確認的正常玩家路徑
# 資料。`stardb.lbx` 是 resolver 測試假檔，故不把不存在的檔案列為完整包前置條件。
moo2_lbx_list="amebafin anatkfin antaroom anwinfin bldg0 bldg1 bldg2 bldg3 bldg4 bldg5 buffer0 cmbtsfx cmbtshp colbldg colgcbt colony colony2 colpups colroads colsum colveggi combat confirm council design dimtvfin diplomat fleet game genwinfn help herodata inbox info intro loserfin mainmenu multigm newgame officer orionfin planets plntdfin plntsum raceopt races racesel science sound starbg stream streamhd techsel turnsum wininfin"

case "${moo2_full_build}" in
0) ;;
1) moo2_app_name+="-full" ;;
*) echo "MOO2_FULL 只能是 0 或 1，得到: ${moo2_full_build}" >&2; exit 2 ;;
esac

if ! docker image inspect "${moo2_image}" >/dev/null 2>&1; then
	echo "找不到既有 macOS 交叉編譯映像: ${moo2_image}" >&2
	echo "請先依 docker/Dockerfile 或既有可重現工具鏈準備它；本腳本不會自行下載或另建映像。" >&2
	exit 1
fi

if [[ "${moo2_full_build}" == "1" ]]; then
	[[ -d "${moo2_data_dir}" ]] || { echo "找不到遊戲資料夾: ${moo2_data_dir}" >&2; exit 1; }
	[[ -f "${moo2_font_file}" ]] || { echo "找不到字型: ${moo2_font_file}" >&2; exit 1; }
fi

for required_dir in "${moo2_dist_dir}" "${moo2_go_cache}"; do
	if [[ ! -d "${required_dir}" ]]; then
		echo "缺少既有目錄: ${required_dir}；拒絕在主機端自行建立。" >&2
		exit 1
	fi
done

moo2_docker_args=(
	docker run --rm --network none --memory 4g --cpus 2 --pids-limit 256
	-u "$(id -u):$(id -g)"
	-e GOPATH=/go -e GOMODCACHE=/go/pkg/mod -e GOCACHE=/go/build-cache
	-e MOO2_APP_NAME="${moo2_app_name}"
	-e MOO2_FULL_BUILD="${moo2_full_build}"
	-e MOO2_LBX_LIST="${moo2_lbx_list}"
	-v "${moo2_repo_root}:/src:ro"
	-v "${moo2_go_cache}:/go"
	-v "${moo2_dist_dir}:/dist"
	-w /src
)
if [[ "${moo2_full_build}" == "1" ]]; then
	moo2_docker_args+=(-v "${moo2_data_dir}:/gamedata:ro" -v "${moo2_font_file}:/font.ttc:ro")
fi

"${moo2_docker_args[@]}" "${moo2_image}" bash -eu -o pipefail -c '
	export PATH=/usr/local/go/bin:$PATH
	app="/tmp/${MOO2_APP_NAME}.app"
	build="/tmp/${MOO2_APP_NAME}-build"
	archive="/dist/${MOO2_APP_NAME}-macos-universal.tar.gz"
	archive_tmp="${archive}.tmp"

	rm -rf "$build" "$app"
	mkdir -p "$build/amd64" "$build/arm64" "$build/universal"

	echo "== [1/5] macOS x86_64 / arm64 交叉編譯 =="
	env GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 CC=o64-clang \
		go build -buildvcs=false -ldflags="-s -w" -o "$build/amd64/moo2" ./cmd/moo2
	env GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 CC=o64-clang \
		go build -buildvcs=false -ldflags="-s -w" -o "$build/amd64/moo2sim" ./cmd/moo2sim
	env GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 CC=oa64-clang \
		go build -buildvcs=false -ldflags="-s -w" -o "$build/arm64/moo2" ./cmd/moo2
	env GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 CC=oa64-clang \
		go build -buildvcs=false -ldflags="-s -w" -o "$build/arm64/moo2sim" ./cmd/moo2sim
	lipo -create -output "$build/universal/moo2" "$build/amd64/moo2" "$build/arm64/moo2"
	lipo -create -output "$build/universal/moo2sim" "$build/amd64/moo2sim" "$build/arm64/moo2sim"
	for bin in "$build/universal/moo2" "$build/universal/moo2sim"; do
		info="$(lipo -info "$bin")"
		echo "$info"
		echo "$info" | grep -q arm64
		echo "$info" | grep -q x86_64
	done

	echo "== [2/5] 組 macOS .app 與 i18n =="
	mkdir -p "$app/Contents/MacOS" "$app/Contents/Resources/assets"
	cp "$build/universal/moo2" "$app/Contents/Resources/moo2-bin"
	cp "$build/universal/moo2sim" "$app/Contents/Resources/moo2sim-bin"
	chmod +x "$app/Contents/Resources/moo2-bin" "$app/Contents/Resources/moo2sim-bin"
	cp -r assets/i18n "$app/Contents/Resources/assets/i18n"

	if [[ "$MOO2_FULL_BUILD" == "1" ]]; then
		echo "== [3/5] 加入本機完整資料子集（不可公開） =="
		res="$app/Contents/Resources"
		mkdir -p "$res/gamedata"
		expected_lbx=0
		for name in $MOO2_LBX_LIST; do
			expected_lbx=$((expected_lbx + 1))
			found=""
			for candidate in "/gamedata/${name}.lbx" "/gamedata/${name}.LBX"; do
				[[ -f "$candidate" ]] && found="$candidate" && break
			done
			if [[ -z "$found" ]]; then
				found="$(find /gamedata -maxdepth 1 -iname "${name}.lbx" -print -quit)"
			fi
			[[ -n "$found" ]] || { echo "缺少完整包必要資料: ${name}.lbx" >&2; exit 1; }
			cp "$found" "$res/gamedata/$(basename "$found" | tr "a-z" "A-Z")"
		done
		[[ "$(find "$res/gamedata" -maxdepth 1 -type f -iname "*.lbx" | wc -l)" == "$expected_lbx" ]]
		cp /font.ttc "$res/font.ttc"
	else
		echo "== [3/5] 驗證公開包不含私有遊戲資料 =="
		if find "$app" -type f \( -iname "*.lbx" -o -iname "font.ttc" \) -print -quit | grep -q .; then
			echo "公開 macOS 包意外包含私有資料或字型" >&2
			exit 1
		fi
	fi

	echo "== [4/5] launcher 與 Info.plist =="
	if [[ "$MOO2_FULL_BUILD" == "1" ]]; then
		launch_args="-game -lang zh -data \"\$RES/gamedata\" -font \"\$RES/font.ttc\""
	else
		launch_args=""
	fi
	cat > "$app/Contents/MacOS/${MOO2_APP_NAME}" <<EOF
#!/bin/bash
DIR="\$(cd "\$(dirname "\${BASH_SOURCE[0]}")" && pwd)"
RES="\$DIR/../Resources"
cd "\$RES"
exec "\$RES/moo2-bin" ${launch_args} "\$@"
EOF
	chmod +x "$app/Contents/MacOS/${MOO2_APP_NAME}"
	cat > "$app/Contents/Info.plist" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict>
  <key>CFBundleName</key><string>${MOO2_APP_NAME}</string>
  <key>CFBundleDisplayName</key><string>銀河霸主2 重製</string>
  <key>CFBundleIdentifier</key><string>com.wicanr2.moo2remake</string>
  <key>CFBundleVersion</key><string>local</string>
  <key>CFBundleShortVersionString</key><string>local</string>
  <key>CFBundlePackageType</key><string>APPL</string>
  <key>CFBundleExecutable</key><string>${MOO2_APP_NAME}</string>
  <key>LSMinimumSystemVersion</key><string>11.0</string>
  <key>NSHighResolutionCapable</key><true/>
</dict></plist>
EOF

	echo "== [5/5] 驗證 bundle 並原子更新 tar.gz =="
	[[ -x "$app/Contents/MacOS/${MOO2_APP_NAME}" ]]
	[[ -d "$app/Contents/Resources/assets/i18n" ]]
	lipo -info "$app/Contents/Resources/moo2-bin"
	lipo -info "$app/Contents/Resources/moo2sim-bin"
	rm -f "$archive_tmp"
	tar czf "$archive_tmp" -C /tmp "${MOO2_APP_NAME}.app"
	# 不可用 `tar ... | grep -q`：在 pipefail 下 grep 提早結束會讓 tar 收到
	# SIGPIPE（141），把已成功的封裝誤判成失敗。
	tar tzf "$archive_tmp" > "$build/archive-members.txt"
	grep -Fxq "${MOO2_APP_NAME}.app/Contents/Resources/moo2-bin" "$build/archive-members.txt"
	rm -rf "/dist/${MOO2_APP_NAME}.app"
	cp -a "$app" "/dist/${MOO2_APP_NAME}.app"
	mv "$archive_tmp" "$archive"
	printf "產出: %s\\n" "$archive"
'

echo "產出: ${moo2_dist_dir}/${moo2_app_name}-macos-universal.tar.gz"
