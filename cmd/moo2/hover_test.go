package main

import "testing"

func TestPointInRectUsesHalfOpenBounds(t *testing.T) {
	if !pointInRect(10, 20, 10, 20, 5, 4) {
		t.Fatal("矩形左上角應命中")
	}
	if pointInRect(15, 23, 10, 20, 5, 4) {
		t.Fatal("矩形右下邊界應採半開區間，不應命中")
	}
}
