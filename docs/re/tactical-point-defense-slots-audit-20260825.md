# 戰術逐槽點防自動開火稽核（2026-08-25）

## 問題

玩家艦已能保存八個武器槽，但飛彈／戰機的點防消費端仍只讀艦艇第一個相容欄位。
因此 PD 位於第二至第八槽時不會自動開火，也無法證明紅色關閉狀態下的原版例外。

## 輸入與工具

- Orion2.exe SHA-256：7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5
- Orion2.exe.i64 SHA-256：4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e
- IDA Pro 9.4，映像 ida-pro-9.4-idapython:py312-v1
- 位址基準：IDA linear，DOS/4GW LE image
- 非破壞性探針：tools/ida/audit_regular_combat_ship_loader.py；原始資料庫唯讀，容器內複本分析。
- 原版內建說明：cmd/moo2/embedded/i18n/help.tsv，SHA-256
  e55e686d0a17e930f7f7d96f693aef24f0047c9c9d049f16ae6b513f1516623b。

## 已證實

1. 內建說明第 402 筆明寫：尚未開火的點防武器，必須在飛彈命中艦艇或戰機接戰前自動開火。
2. 內建說明第 627 筆明寫：紅色表示武器關閉，但點防武器仍會在飛彈命中前開火。
3. 外部索引的 Load_Combat_Ship_ 導覽標籤對應原始 sub_4954A @ 0x4954A，邊界
   0x4954A..0x49A41，直接 caller 為 0x42868、0x47C0E、0x48450。
4. 0x4964A 將槽索引清為零；0x49663 以 0x0B 為 runtime 槽 stride，從 129-byte
   設計 record 的 +0x1C..+0x23 複製到 313-byte 戰鬥 record 的 +0x52..+0x58。
   0x49743..0x4974D 遞增槽索引並以小於 8 回迴，因此原版 runtime 不是單武器欄位。
5. 0x49744 在載入每槽時寫入 runtime slot+0x0A，即艦艇基址 +0x5C = 0。
   這與 Draw_Weapon_Status_Display_ @ 0x2E2CD 讀狀態及 Do_Combat_Turn_ @ 0x42F7F
   寫回待命狀態的既有證據共同證明「狀態是逐槽」。

## 強推論與未知

- 已證實玩家可見契約：自動點防不得被紅色主動開火狀態阻擋。
- 強推論：同一槽的 WorkingCount 代表同型武器門數，自動觸發時每門各開火一次；
  這與原版設計 record 的數量欄位及說明使用複數 weapons 一致，但本輪未追完自動點防的逐門迴圈。
- 未知：多槽 PD 之間的原版確切亂數消費順序、飛彈已全數攔截後是否仍讓尾槽開火，以及一般 AI
  艦隊 blueprint 對應到 remake 戰略艦隊的完整來源。

## Remake 對映

- 自動點防依 WeaponMounts 槽序找所有 typed PD，不讀 WeaponModes。
- 每槽一回合只自動觸發一次；同槽 WorkingCount 依序開火，並共用艦級的攔截餘數。
- 舊存檔／單槽艦繼續使用 WeaponName、Mods 與 PointDefenseSpent 相容路徑。
- 快速結算與格子戰術都消費逐槽 PD；只有格子戰術有紅／橘／綠狀態，且自動點防刻意忽略該狀態。

