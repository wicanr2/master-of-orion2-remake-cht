# 多槽艦艇 blueprint 過渡規格

## 目的

先停止 `.GAM` 與自動設計資料在匯入 runtime 時被壓成第一個武器／第一個特殊裝置。這是把
`Auto_Design_Ship_` 接到原版多槽資料模型的必要前置，不代表兩條戰鬥路徑已完成多槽消費。

## Typed runtime 欄位

`Ship.WeaponMounts` 保存最多八筆：

- 原版 raw weapon ID（未知時仍保留）
- 顯示名稱
- 最大／可用數量
- 火線角、raw mods、彈藥
- 單發最大攻擊值

`Ship.SpecialIDs` 保存原版 bitset 中所有設定位元的 raw ID；`Ship.Special` 暫保留第一筆的
舊 UI／戰鬥相容名稱。

## 寫入端

- `.GAM` importer 必須走完八個 weapon records 與全部 special bits，不可遇到第一筆就 return。
- remake 單槽造艦也要寫入一筆 `WeaponMounts`，使新舊來源共用 typed 欄位。
- JSON／熱座／TCP 因 `Ship` 已在 `Fleet` 快照中，新增 slice 必須自然往返且有測試。

## 相容欄位

在所有多槽 consumer 完成前：

- `Weapon`／`WeaponAttack`／`Arc`／`WeaponAmmo` 對映第一個有效 mount。
- `Special` 對映第一個 special ID 的顯示名稱。
- 不把其餘槽偷偷加總進舊戰鬥公式；那會造成只有部分路徑改變的半套行為。

## 後續 gate

多槽模型只有在快速結算、格子戰術、設計 UI、建造／改裝成本、損傷、修復與存檔都逐槽消費後，
才可把 `Auto_Design_Ship_` 標為完整接線。
