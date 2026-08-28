# 混合種族 colonist 產出與成長稽核（2026-08-25）

## 證據身分

- 輸入：`Orion2.exe`
- SHA-256：`7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`
- 工具：IDA Pro 9.4、`tools/ida/audit_colonist_production.py`
- 位址空間：IDA linear；DOS/4GW LE object #1
- 方法：唯讀 `.i64` 交叉參照與原始指令匯出；沒有改名或寫回資料庫。

## 已證實

1. `sub_DE280 @ 0xDE280` 對每筆 packed colonist 取 `raw & 0x0F`，並把該值傳給
   `sub_DE22C @ 0xDE22C` 與 `sub_DDF2C @ 0xDDF2C`。這個值是當局 player slot，
   不是十三個內建種族的固定物種索引；一般帝國為 0..7，8／9 是 Android／Natives。
2. 食物 `sub_DE0C6 @ 0xDE0C6`、工業 `sub_DED47 @ 0xDED47`、研究
   `sub_DFE77 @ 0xDFE77` 分別用 colonist slot 乘 `0xEA9`，讀該 player runtime 的
   `+0x8A1`／`+0x8A2`／`+0x8A3`。重力 helper 以同一 slot 讀 `+0x8A9` Low-G 與
   `+0x8AA` High-G。因此混居殖民地必須逐群保存來源 slot 與其 runtime trait profile。
3. `sub_DE0C6` 另外讀 colonist slot 的 `+0x8AB` Aquatic，依殖民地氣候重算食物基礎；
   不能只用「殖民地所有者已烘入的 FoodPerFarmer」替全部人口計算。
4. `Apply_Colony_Pop_Growth_`（`sub_E2DCA @ 0xE2DCA`）先按各 race slot 的
   `colony+0xB4[slot]` 成長池選一個 slot，達 1000 後新增 packed colonist，將低四位寫成
   被選中的 slot，loyalty 寫 colony owner，job 另行選定。成長人口不會無條件變成所有者種族。
5. 同函式在某 race slot 人口過量／成長池為負時，會從該 slot 的 packed colonist 候選中刪除
   一筆並以最後一筆回填；這再次證明人口變動必須維持 race 分布，而不只是三職務總數。
6. 特殊 slot 的三項 helper 分支已閉合：
   - Android（slot 8）：`0xDE1DD..0xDE1E4` 食物 `+6`、`0xDEDE3..0xDEDE5`
     工業 `+3`、`0xDFF33..0xDFF35` 研究 `+3`。
   - Natives（slot 9）：`0xDE1E6..0xDE1EC` 食物 `+4`；工業與研究沒有額外值。
   - `sub_DDF2C @ 0xDDF2C` 在 `0xDDF3B..0xDDF3F` 對 slot >= 8 直接走回傳 4，
     即完全不受 Low／Normal／Heavy-G 懲罰。
7. `sub_E1839 @ 0xE1839` 會先清空 `colony+0xC8[10]`，再逐 slot 計算成長率；
   `0xE1CA4..0xE1CCE` 的核心與官方手冊公式一致：以該 slot 人口、殖民地總人口及該
   race 的人口上限形成平方根基礎值，再乘該 race 的人口成長 trait／殖民地修正。
   `sub_E2DCA` 將這個 signed rate 加入 `colony+0xB4[slot]`，達 1000 才新增該 slot colonist。
8. `sub_E3456 @ 0xE3456` 的舊導覽名「Compute Colony Pop Growth」錯誤。它只在
   `colony+0x12F == 3` 時累積 `+0x12E`，每 240 點挑一名 prisoner 清 bit 10／改 loyalty；
   是同化 consumer，不是一般人口成長公式。受版控腳本已改稱
   `raw_Apply_Colonist_Assimilation`，保留原函式名與位址。
9. 勘誤：人口稽核腳本一度把 `0xE2000` 加成「人口點數重算」根節點；這與獨立維護費切片
   已證實的 `Compute_Player_Maintenance_ @ 0xE2000` 衝突，而且逐槽成長鏈不需要該入口。
   該錯誤根節點已移除；既有已證實人口結論均來自 `sub_E1839`／`sub_E2DCA`，不受影響。
10. `sub_DEB4B @ 0xDEB4B` 逐 packed colonist 重建食物消耗。一般人口預設每人寫入 2 個
    半食物單位；player `+0x8B0`（Cybernetic）改為 1，`+0x8B1`（Lithovore）改為 0。
    owner、外族、prisoner、Natives 分別累積至 colony `+0xFC..+0xFF`；Android 不吃食物，
    Natives 固定每人 2。每個一般 player slot 的消耗另存 `+0x104[slot]`。
11. `sub_DF546 @ 0xDF546` 以相同分類重建工業消耗。一般 player slot 只有 `+0x8B0`
    Cybernetic 每人消耗 1 個半工業單位；Android 固定每人消耗 2，Natives 為 0。
    owner／Android／外族／prisoner 分別寫 `+0x100..+0x103`，逐 slot 值寫 `+0x10C[slot]`。
    `+0xF0=(四類半工業總和+1)/2`；它不在 `sub_DEE1B` 早扣，而由
    `Apply_Production_ @ 0xE3806..0xE382A` 以 `max(colony+0xE9-colony+0xF0,0)` 減少建造進度。
    完整外層見 `colony-industry-production-pollution-audit-20260828.md`。
12. `sub_E1839 @ 0xE192B..0xE1A8C` 依固定優先序分配可用資源：食物依 owner → 外族 →
    prisoner → Natives；工業依 owner → Android（一次必須供足 2）→ 外族 → prisoner。
    未滿足的 owner 食物／工業與 Natives 食物每半單位各形成 `-25` 成長點；Android 每個
    未滿足的半工業需求形成 `-500`。外族與 prisoner 的總短缺先乘 `-25`，再依各非 owner
    slot 的原始食物＋工業需求比例分配。這些負值先寫入 `+0xC8[slot]`，之後才與正常
    平方根成長值相加，因此饑荒不是「停止正成長」，而是正負項可互相抵銷。
13. `sub_E2DCA @ 0xE2F24..0xE309A` 先選一個必須保留一人的 slot：若 owner 尚有人就保護
    owner；否則保護一般 player slot 中人口最多者（同數時較高 slot），再回退 Android、
    Natives。每個 slot 累積池低於 0 時，只要人口多於該保留量便刪一人並加回 1000 點。
14. 刪除候選只限相同 slot。`0xE3030` 的 job mask `0x180` 使工人／科學家優先於農夫；
    每個候選都呼叫 `sub_1247A0(count)`，回 1 時替換目前候選，形成 reservoir sampling。
    選定後用 packed colonist 最末筆回填洞。remake 聚合群組不保存原陣列順序，但可保留
    同一候選集合、均勻機率、職務優先與可存檔亂數流；逐位元 PRNG／排列順序不宣稱 parity。
15. `sub_E2DCA` 先在 `0xE2F87..0xE30A4` 完成全部 slot 的負池刪除，之後才於
    `0xE313D..0xE3444` 處理正池新增；兩階段不可交錯。正成長陣列只建立一般 player slots，
    不包含 Android／Natives，因此兩種特殊人口沒有自然正成長。
16. 正成長先把 colony owner 交換到第 0 位，`sub_FE9F5 @ 0xFE9F5` 只洗牌其餘 entries。
    該 helper 對 `i=0..n-1` 呼叫 `sub_1247A0(n-i)` 取得 1-based 位移並交換 `i` 與
    `i+roll-1`，是包含最後一格 `Random(1)` 的 Fisher–Yates。這個順序會在多 slot 同時跨過
    1000 且總人口容量有限時決定誰先新增。

## 強推論與實作邊界

- 原版 `Colony.FoodPerFarmer/ResearchPerScientist` 是 owner 快取捷徑的高層解釋仍屬強推論；
  `IndustryPerWorker` 的 environment／building／technology base、`colony+0xDE` writer 與
  非 owner 重算路徑已由
  [`colony-industry-per-worker-audit-20260828.md`](colony-industry-per-worker-audit-20260828.md)
  閉合。remake 可保留既有快取作
  舊 JSON fallback，但 typed groups 生效時必須把 owner 種族修正拆出後換成群組修正。
- `sub_E1839` 的住房、科技、Medicine 領袖、Cloning Center、事件、AI 難度、prisoner 衰減
  與完整正 rate 已於 2026-08-28 由 producer／consumer 閉合；`sub_E2DCA` 的滅絕與新人口
  職務回寫亦已補完。見
  [`population-growth-runtime-audit-20260828.md`](population-growth-runtime-audit-20260828.md)。
  remake 仍只能把已具 typed model 的部分接入，不因公式已知就宣稱現行資料模型完整對齊。
- 原版 packed colonist 的實際排列沒有保存在 remake typed groups，因此雖可重現 reservoir
  sampling 的分布與抽取流程，無法保證同一原版 save／PRNG 狀態會刪到同一筆 packed record。

## 推翻的舊模型

`ColonyState.RaceGravity` 只能描述 owner／單一種族殖民地。它仍是舊 JSON fallback，不能作為
混居殖民地的最終資料來源。把 packed 低四位直接當 `Race.OrigIdx` 同樣錯誤：兩者位址空間不同。
