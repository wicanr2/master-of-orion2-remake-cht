# 母星、重力、研究與 Warlord traits 稽核（2026-08-28）

## 證據契約

- 輸入：`Orion2.exe`，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- IDA 資料庫：`Orion2.exe.i64`，SHA-256
  `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`。
- 工具：IDA Pro 9.4、`tools/ida/audit_custom_race_trait_consumers.py`，並交叉引用既有研究、
  艦員、地面戰與重力稽核；位址均為 IDA linear、DOS/4GW LE object #1。
- 可重生證據：
  [`evidence/custom-race-trait-consumers-ida-20260828.json`](evidence/custom-race-trait-consumers-ida-20260828.json)。
  外部符號只供導覽，不取代 raw 位址、bytes 與 operand。

## 母星 traits：`+0x8AD..+0x8AF`

`Modify_Home_Worlds_ @ 0x7C4AF` 逐玩家選定母星後直接修改 planet／star record：

| trait | runtime | 原版寫入 |
| --- | --- | --- |
| Large Homeworld | `player+0x8AD != 0` | planet size byte `planet+0x05 = 3`，即 Large。 |
| Poor Homeworld | signed `player+0x8AE < 0` | 呼叫礦產 setter，目標 raw richness `1`。 |
| Rich Homeworld | signed `player+0x8AE > 0` | 呼叫同一 setter，目標 raw richness `3`。 |
| Artifacts Homeworld | `player+0x8AF != 0` | `planet+0x0F = 10`，並同步 planet record 的 special 欄；raw 10 已由研究 consumer 證實為 Ancient Artifacts。 |

`+0x8AE` 是單一 signed byte 的 `-1/0/+1`，不是 Rich 與 Poor 兩個旗標。Large raw 3 與
礦產 raw 1／3 均落在既有 planet enum；一般中值保持生成器結果，不由本函式硬寫另一常數。

`Twiddle_Selected_Adv_Civ_Planets_ @ 0x638A9` 在 Advanced Civilization 開局會把 Artifacts
trait 納入高價 special 的重新分配／平衡流程；它不改變上述標準母星 raw 10 契約。2026-08-30
已閉合其額度、距離／owner gate、worth 排序、輪流選取、最高玩家 90% 平衡、每顆六次升級與
special 4／5／10 再分配，見
[`advanced-civilization-planets-audit-20260830.md`](advanced-civilization-planets-audit-20260830.md)。

## Low-G／High-G：`+0x8A9..+0x8AA`

### 母星與殖民產出

`Modify_Home_Worlds_` 將 Low-G、Normal-G、High-G 母星分別交給 `Enforce_Gravity_ @ 0x7B45C`
寫 raw gravity `0/1/2`。`Gravity_Player_Production_Factor_ @ 0xDDF2C` 對每名 colonist race
判定，優先序是 Planetary Gravity Generator → High-G → Low-G → Normal-G；回傳值映成
產出 `0%/-25%/-50%` 懲罰。混合人口必須逐 race slot 計算，不能只看殖民地 owner。

### 地面戰與軌道轟炸

- Low-G 在地面戰共同攻擊 block 加固定 `-10`；不是「乘 0.9」。
- High-G 令四種地面單位基礎耐受由 1 變 2。
- `Resolve_Bomb_Hit_ @ 0xDCEBD` 對 High-G 玩家的人類陸戰隊與裝甲各把消滅門檻再加 1；
  Low-G 沒有對稱的直接轟炸門檻分支。
- AI 殖民地重力評分表已證實 High-G 只懲罰 Low-G world、Low-G 分別懲罰 Normal／Heavy world；
  這是 AI 選址分數，不應重複套進玩家產出。

## Uncreative／Creative：`+0x8B4..+0x8B5`

- 玩家與 AI 都在開始投入研究前保存 field `player+0x321` 與 application `+0x322`。
- Creative 在 `sub_E4410 @ 0xE4410` 突破時掃描 field application table，授予全部 status 1
  applications；field 22、23、28、29、55、57 本來就固定全授予，不依 Creative。
- Uncreative 在 `sub_E4204 → sub_E408F` 形成下一個 field 的可選集合時，以
  `Random(candidateCount)` 的 reservoir-style 選擇只留下單一 status 1 application。
- 因此 Uncreative 的亂數發生在可選集合形成，不是突破後從完整清單臨時抽一個；Creative
  也不是把一般玩家的 pending choice 全部視為擁有。

## Warlord `+0x8BD`

### 艦員與領袖等級

`Calc_Ship_Level_ @ 0x147E7` 依 CrewXP 50／150／500／1000 門檻計 raw level，Warlord 再
加一並夾到 5；因正常每回合 XP 夾 500，一般最高 Elite，Warlord 可達 Ultra Elite。
`Owned_Officer_Level_ @ 0x94951` 把同一 raw trait 傳入 level helper；leader ID `0x42`
固定忽略 Warlord。`Increase_Officer_Level_By_One_` 亦保留該加級旗標，證明這不是純顯示偏移。

### 指揮點與地面部隊容量

`Compute_Player_Maintenance_ @ 0xE2000` 在逐殖民地 command contribution 分支中，Warlord
每座符合條件的殖民地把 local accumulator 加 2；它不是全帝國最後只加一次 2。

四個 ground-unit limit helper 皆採「無 Warlord 再除 2」：Marine 上限為人口／容量基值，
Armor 先有自身 `/2`，再視 Warlord 決定是否額外 `/2`。因此 Warlord 對兵營容量是兩倍，
且結果最低 1；`Produce_Ground_Military_ @ 0xE3616` 每五回合生產與超額逐回合裁減都消費
這些上限。

## 閉合結論與剩餘邊界

- **已證實**：三項母星 raw 寫入、signed Rich／Poor 共欄、三種母星重力、逐 race 產出、
  Low-G 地戰 -10、High-G 耐受與轟炸門檻、Creative／Uncreative 時序、Warlord 艦員／領袖、
  每殖民地 command +2 與地面部隊容量。
- **仍待獨立切片**：重力轟炸的完整武器強度上游。母星 traits 與 Warlord 的 NPC raw profile
  權重已於 2026-08-30 由完整 direct-site 表閉合；raw 候選正式名稱仍未知，但不再把數值權重
  列為缺口。這些不推翻已閉合玩家規則，也不能由鄰近立即數猜填。
- C runtime、Watcom helper、Random 內部算法及平台 API 不納入玩法分母。
