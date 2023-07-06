package main

func isWithinBoard(x, y int) bool {
	if x >= 0 && x < gameGridWidth && y >= 0 && y < gameGridHeight {
		return true
	}
	return false
}

// getTopOffset - get the top offset of a shape
func getTopOffset(shape int) int {
	for i := 3; i >= 0; i-- {
		if (shape>>(i*4))&0xF != 0 {
			return 3 - i
		}
	}
	return 0
}

func getShapeDimensions(shape int) int {
	minX, maxX := 4, 0
	for i := 0; i < 16; i++ {
		if shape&(1<<uint(i)) != 0 {
			x := i % 4
			if x < minX {
				minX = x
			}
			if x > maxX {
				maxX = x
			}
		}
	}
	width := maxX - minX + 1
	return width
}

func findFirstNonEmptyColumn(shape int) int {
	columnIndex := 4
	for i := 0; i < 16; i++ {
		if shape&(1<<uint(15-i)) != 0 {
			col := i % 4
			if columnIndex > col {
				columnIndex = col
			}
		}
	}
	return columnIndex
}
