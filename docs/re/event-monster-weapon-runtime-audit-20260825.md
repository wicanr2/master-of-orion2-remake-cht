# 事件怪物武器執行期稽核（2026-08-25）

## 證據契約

- 輸入：`Orion2.exe`，SHA-256 `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- 工具：IDA Pro 9.4；位址為 IDA linear、DOS/4GW LE object #1。
- 唯讀匯出器：`tools/ida/audit_monster_weapon_runtime.py`；本輪輸出
  `/tmp/monster-weapon-runtime-v2.json`。未改名、未套型別、未寫回原始資料庫。

## 已證實

`Load_Combat_Ship_ @ 0x4954A` 把每個設計槽的 raw mods `ShipDesign +0x21` 複製到 11-byte
combat weapon record 的 `+0x56`。以下兩個消費端直接閉合低位旗標：

- `sub_2B9E3 @ 0x2B9E3`：bit `0x0004` 先把武器表傷害除二；bit `0x0002` 再乘 3/2。
- `sub_39434 @ 0x39434`：bit `0x0002` 把射程參數除二並走 +50 分支；bit `0x0004` 把射程
  參數加倍（上限 8），另走 +25／-50 分支。

這與原版手冊 Heavy Mount「150% 傷害、射程懲罰減半」及 Point Defense「半傷害、射程懲罰
加倍、命中 +25」逐項相符，因此：

| raw mask | 語意 | 證據等級 |
|---:|---|---|
| `0x0002` | Heavy Mount | 已證實 |
| `0x0004` | Point Defense | 已證實 |

因此 Crystal／Hydra 的主武器是 Heavy Mount；Dragon 的 20 門 ID 41 是 Point Defense。
武器 ID、類別、傷害與數量則由 loader 與 `0x17F807` 的 46 筆武器表共同證實。

## 後續勘誤與停止線

- 本輪原先把 Dragon `0x4000` 留為未知；後續追到 `sub_3CEB7`／`sub_3D2DF` 的 +50 ordnance
  消費端、category 2 改造表與 ID 40 每格 -15 分支，已證實為 OVR，不是 NR。完整勘誤見
  [`event-monster-dragon-raw4000-audit-20260825.md`](event-monster-dragon-raw4000-audit-20260825.md)。
- 怪物專用武器 ID 44／45 的特殊 runtime 亦已在後續切片閉合；本文件保留當時證據邊界，
  現況以 special runtime 稽核及 WORKLIST 活表為準。
- 本輪只把已證實的 mount 數量、武器表傷害、HV／PD 接到快速反擊；戰術 sprite、逐格射程、
  特殊動畫與選目標仍是後續切片。
