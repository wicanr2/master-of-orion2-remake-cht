# ebiten 移植工程筆記(Phase 2)

> 記錄 go/ebiten backend 的實作要點、docker headless 設定與踩到的事實。程式:`cmd/moo2`、`docker/Dockerfile.ebiten`、`scripts/screenshot.sh`。

## 1. 關鍵事實:MOO2 是 640×480

MOO2 畫面為 **640×480**(openorion2 `screen.h` 亦以此為邏輯座標),**不是** 320×200。
- 影響:CJK 中文渲染的空間比 320×200 遊戲寬鬆;kickoff 文件早期「320×200 底圖」假設已據此修正。
- 實作:`Layout` 回傳背景圖實際 bounds(主選單 = 640×480),window 設為同尺寸。

## 2. 資料層 → ebiten 全鏈路(已驗證)

```
assets.Resolver.OpenLBX → lbx.DecodeImage → Frame.ToRGBA(palette,keycolor)
  → ebiten.NewImageFromImage(*image.RGBA) → screen.DrawImage
```

`MAINMENU.LBX` 資產 21(`ASSET_MENU_BACKGROUND`)在 ebiten 下渲染出完整正確的主選單(標題/選單鈕/太空船/星雲),顏色與原版一致 → 整條解碼+繪製管線像素級正確。

## 3. Docker headless 設定(對齊 mom playbook §6)

- **ebiten 需 CGO**(Linux 走 X11/OpenGL):`Dockerfile.ebiten` 裝 `gcc pkg-config libgl1-mesa-dev libx{randr,cursor,inerama,i}-dev libxxf86vm-dev`,`CGO_ENABLED=1`。
- **headless 繪圖**:`xvfb` + `xauth`(xvfb-run 需要)+ `libgl1-mesa-dri` 軟體渲染,`LIBGL_ALWAYS_SOFTWARE=1`。
- **截圖**:不依賴 WM/imagemagick —— app 內 `screen.ReadPixels` 讀回像素直接存 PNG(`-shot` 模式跑 N 幀後結束)。比 `import -window root` 更確定性。
- **`-buildvcs=false`**:容器內 mount 的 `.git` 擁有權與容器 user 不同 → `go build` 對 main 套件做 VCS 戳記會 `exit 128`,關掉即可。

## 4. 對映 openorion2 架構(見 kickoff/06)

- openorion2 `Screen` 抽象介面 → 之後以 ebiten backend 實作(目前 `cmd/moo2` 是直接繪製的骨架,尚未抽介面)。
- 事件:openorion2 只有滑鼠;ebiten 走 `Update/Draw`,鍵盤另補。

## 5. 待續

- [ ] 把 openorion2 `Screen` 介面以 ebiten 實作(registerTexture/drawTexture/fillRect/clip)。
- [ ] 載入存檔 → 繪製星系圖(M2 完整里程碑;目前已能顯示背景圖)。
- [ ] 資產快取(避免每幀 NewImageFromImage)。
