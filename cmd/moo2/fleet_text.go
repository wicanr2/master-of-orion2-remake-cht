package main

const (
	fleetRosterX     = 20
	fleetRosterW     = 304
	fleetRosterRowH  = 18
	fleetRosterFont  = 9
	fleetRosterStart = 300
)

func fleetHeaderTextRect(y int) textSafeRect {
	return textSafeRect{x: fleetRosterX, y: y - 9, w: fleetRosterW, h: fleetRosterRowH, insetX: 4, insetY: 1}
}

func fleetSplitTextRect(y int) textSafeRect {
	return textSafeRect{x: 28, y: y - 9, w: 296, h: fleetRosterRowH, insetX: 4, insetY: 1}
}

func fleetShipNameTextRect(y int) textSafeRect {
	return textSafeRect{x: 34, y: y - 9, w: 106, h: fleetRosterRowH, insetX: 4, insetY: 1}
}

func fleetShipClassTextRect(y int) textSafeRect {
	return textSafeRect{x: 140, y: y - 9, w: 104, h: fleetRosterRowH, insetX: 2, insetY: 1}
}

func fleetShipDamageTextRect(y int) textSafeRect {
	return textSafeRect{x: 244, y: y - 9, w: 80, h: fleetRosterRowH, insetX: 2, insetY: 1}
}
