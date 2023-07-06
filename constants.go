package main

// Tetris pieces can have four rotations and there are seven unique shapes I, O, T, S, Z, J, and L total of 28 entries
// the shapes are represented as a 4x4 grid of bits
// array contains 30 entries: it includes the 7 shapes each rotated in 4 ways,
// with 2 duplicates for the "I" shape and 1 duplicate for the "O" shape, for a total of 30
var shapes = [][]int{
	// I
	{0x0F00, 0x2222, 0x0F00, 0x2222},
	// O
	{0x6600, 0x6600, 0x6600, 0x6600},
	// T
	{0x4C40, 0x4E00, 0xC880, 0xE400},
	// S
	{0x06C0, 0x8C40, 0x6C00, 0x4620},
	// Z
	{0x0C60, 0x4C80, 0xC600, 0x2640},
	// J
	{0x44C0, 0x8E00, 0xE880, 0xC440},
	// L
	{0x4460, 0x0E80, 0xC440, 0x2E00},
}

var menu = []string{
	"Tetris game",
	"",
	"left   Left",
	"right  Right",
	"up     Rotate",
	"down   Down",
	"esc,q  Exit",
}

const (
	gameGridWidth  = 10
	gameGridHeight = 20
)

type ShapeState int

const (
	EmptyState ShapeState = iota
	FallingState
	FixedState
)
