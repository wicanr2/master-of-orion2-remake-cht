package shell

// ground_types.go:地面戰的部隊類型索引(對應 `gamedata.GroundSide` 的四格)。
//
// 原版有四種(`Ground_Combat_Round_` 的 `cmp byte ptr [ebx+16h], 4`),但**沒有對出名字**。
// remake 目前只用得到兩種,所以只給這兩個常數——
// 編一個「類型 2 = ???」的名字比留白更糟。
const (
	groundTypeMarines = 0 // 陸戰隊
	groundTypeTanks   = 1 // 戰車營
)
