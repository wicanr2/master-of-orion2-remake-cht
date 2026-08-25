# 所有者種族重力產出稽核（2026-08-25）

## 證據身分

- 輸入：`Orion2.exe`
- SHA-256：`7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`
- 工具：IDA Pro 9.4、`tools/ida/audit_colonist_production.py`
- 位址空間：IDA linear；DOS/4GW LE object #1
- 方法：唯讀 `.i64` 交叉參照與原始指令匯出；未改名、未寫回資料庫。

## 已證實

1. `sub_DDF2C @ 0xDDF2C` 接收 colony 與 colonist race。若殖民地有
   Planetary Gravity Generator（colony `+0x14F`），直接回傳 4。
2. 否則函式讀 planet gravity，以及該 colonist race 的 player runtime：
   `+0x8AA` 先判 High-G、`+0x8A9` 再判 Low-G。這兩格與受版控的
   `TRAIT_HIGH_G`／`TRAIT_LOW_G` 寫入與其他消費端一致。
3. 回傳值只有 4、3、2。`sub_DE280 @ 0xDE280` 把 `4-return` 乘上
   `base * -5`，故三值分別等價 0%、-25%、-50%；`sub_DDFD3 @ 0xDDFD3`
   也把同一值乘入逐人口產出。
4. 判定優先序是 Planetary Gravity Generator → High-G → Low-G → Normal-G。
   因此同時出現兩個互斥 trait 的髒資料時，原版採 High-G。

## 實作邊界

本切片先讓單一種族／所有者人口使用正確重力；原版真正以每名 colonist race 查值。
混合種族殖民地仍需 typed colonist group，不能把 owner 值宣稱成完整 parity。
