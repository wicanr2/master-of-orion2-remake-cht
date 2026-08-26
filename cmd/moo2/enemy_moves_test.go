package main

import (
	"math"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

func nearFloat32(a, b float32) bool {
	return math.Abs(float64(a-b)) < 0.01
}

func TestEnemyMoveGeometryProducesRouteAndMovingMarker(t *testing.T) {
	stars := []shell.Star{{X: 0.1, Y: 0.2}, {X: 0.8, Y: 0.7}}
	move := shell.EnemyFleetMove{
		AIIndex: 0, FromStar: 0, ToStar: 1, ETA: 2,
	}
	x1, y1, x2, y2, mx0, my0, ok := enemyMoveGeometry(stars, move, 0)
	if !ok {
		t.Fatal("有效 Enemy Moves 航線應產生幾何資料")
	}
	if !nearFloat32(mx0, x1) || !nearFloat32(my0, y1) {
		t.Fatalf("tick 0 的 marker 應從起點開始：start=(%.2f,%.2f) marker=(%.2f,%.2f)", x1, y1, mx0, my0)
	}
	_, _, _, _, mx1, my1, ok := enemyMoveGeometry(stars, move, 45)
	if !ok || !(mx1 > min(x1, x2) && mx1 < max(x1, x2)) || !(my1 > min(y1, y2) && my1 < max(y1, y2)) {
		t.Fatalf("中段 tick 的 marker 應在線段內：line=(%.2f,%.2f)-(%.2f,%.2f) marker=(%.2f,%.2f)", x1, y1, x2, y2, mx1, my1)
	}
}

func TestEnemyMoveGeometryRejectsInvalidIndices(t *testing.T) {
	_, _, _, _, _, _, ok := enemyMoveGeometry([]shell.Star{{X: 0.5, Y: 0.5}}, shell.EnemyFleetMove{
		AIIndex: 0, FromStar: 0, ToStar: 4, ETA: 2,
	}, 0)
	if ok {
		t.Fatal("越界航線不得產生繪製幾何")
	}
}
