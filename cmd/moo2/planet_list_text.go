package main

func planetListColumnRect(row, column int, secondary bool) textSafeRect {
	left := [...]int{18, 97, 176, 260, 345}
	width := [...]int{78, 78, 82, 84, 74}
	cy := int(planetListRowY[row])
	if secondary {
		return textSafeRect{x: left[column], y: cy + 3, w: width[column], h: 21, insetX: 2, insetY: 1}
	}
	return textSafeRect{x: left[column], y: cy - 22, w: width[column], h: 27, insetX: 2, insetY: 1}
}

func planetListEmptyTextRect() textSafeRect {
	return textSafeRect{x: 16, y: 38, w: 398, h: 46, insetX: 4, insetY: 2, lineH: 20}
}

func planetListMessageTextRect() textSafeRect {
	return textSafeRect{x: 454, y: 354, w: 157, h: 30, insetX: 4, insetY: 2, lineH: 13}
}
