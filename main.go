package main

import (
	"context"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gdamore/tcell"
)

// Game represents the game state which includes the game screen,
// the current active piece, the game board, and a context for managing
// the game's lifecycle. Also it includes a mutex to ensure thread-safety
// during game operations
type Game struct {
	screen tcell.Screen
	piece  *Piece
	board  Board
	ctx    context.Context
	cancel context.CancelFunc
	lock   sync.Mutex
}

// Shape represents a game piece's shape, its rotation state, and current
// state whether it's falling or fixed
type Shape struct {
	shape int
	pos   int
	state ShapeState
}

// Position represents the position of a game piece on the board.
type Position struct {
	x, y, oldX, oldY int
}

// Piece represents a game piece including its shape and its position on the board.
type Piece struct {
	Shape         Shape
	Position      Position
	PreviousShape Shape
}

// Board represents the game board which is a 2D array of ShapeState.
type Board [gameGridHeight][gameGridWidth]ShapeState

// newGame initializes a new game with a given screen, context and cancellation function.
// It creates a new game piece and returns a pointer to the newly created game.
func newGame(screen tcell.Screen, ctx context.Context, cancel context.CancelFunc) *Game {
	g := &Game{
		screen: screen,
		ctx:    ctx,
		cancel: cancel,
	}
	g.newPiece()
	return g
}

// newPiece creates a new game piece and sets it as the current piece of the game
func (g *Game) newPiece() {
	shape, pos := rand.Intn(1), rand.Intn(4)
	x, y := -findFirstNonEmptyColumn(shapes[shape][pos]), -getTopOffset(shapes[shape][pos])
	g.piece = &Piece{
		Shape: Shape{
			shape: shape,
			state: FallingState,
			pos:   pos,
		},
		Position: Position{
			x:    x,
			y:    y,
			oldX: x,
			oldY: y,
		},
		PreviousShape: Shape{
			shape: shape,
			pos:   pos,
		},
	}
}

// handleKey handles key events for the game. It enables control of the game
// piece's movements using keyboard
func (g *Game) handleKey(ev *tcell.EventKey) {
	// control the piece movements using keyboard
	switch ev.Key() {
	case tcell.KeyEsc, tcell.KeyCtrlC:
		g.cancel()
		return
	case tcell.KeyLeft:
		g.moveLeft()
	case tcell.KeyRight:
		g.moveRight()
	case tcell.KeyDown:
		g.moveDown()
	case tcell.KeyUp:
		g.rotate()
	}

	if ev.Rune() == 'q' || ev.Rune() == 'Q' {
		g.cancel()
	}
}

// moveShape moves the current game piece to a new position on the board
func (g *Game) moveShape() {
	// remove the old shape from the board
	oldShape := shapes[g.piece.PreviousShape.shape][g.piece.PreviousShape.pos]
	g.fixPiece(oldShape, g.piece.Position.oldX, g.piece.Position.oldY, 0)
	// place the new shape on the board
	current := shapes[g.piece.Shape.shape][g.piece.Shape.pos]
	g.fixPiece(current, g.piece.Position.x, g.piece.Position.y, g.piece.Shape.state)
	// update the piece position
	g.piece.Position.oldX = g.piece.Position.x
	g.piece.Position.oldY = g.piece.Position.y
	g.piece.PreviousShape.shape = g.piece.Shape.shape
	g.piece.PreviousShape.pos = g.piece.Shape.pos
}

// drawBoard draws the game board on the screen. It sets the color of
// the blocks depending on whether they are part of a shape or just the background
func (g *Game) drawBoard() {
	shapeColor := tcell.StyleDefault.Foreground(tcell.ColorBlue)
	bg := tcell.StyleDefault.Foreground(tcell.ColorLightCyan)

	for y := 0; y < gameGridHeight; y++ {
		for x := 0; x < gameGridWidth; x++ {
			g.screen.SetContent(x, y, '#', nil, bg)
		}
	}

	for y := 0; y < gameGridHeight; y++ {
		for x := 0; x < gameGridWidth; x++ {
			if g.board[y][x] != 0 {
				g.screen.SetContent(x, y, '#', nil, shapeColor)
			}
		}
	}
}

// fixPiece updates the board with the given shape at the given position
// with the specified state. It helps in moving a shape around the board
func (g *Game) fixPiece(shape, x, y int, val ShapeState) {
	for i := 0; i < 16; i++ {
		if shape&(1<<uint(15-i)) != 0 {
			xx := x + i%4
			yy := y + i/4
			if xx >= 0 && xx < gameGridWidth && yy >= 0 && yy < gameGridHeight {
				g.board[yy][xx] = val
			}
		}
	}
}

// hasCollision checks if the piece in the given shape and position
// would collide with the wall or another piece on the board
func (g *Game) hasCollision(x, y, pos int) bool {
	shape := shapes[g.piece.Shape.shape][pos]
	for i := 0; i < 16; i++ {
		if shape&(1<<uint(15-i)) != 0 {
			xx, yy := x+i%4, y+i/4
			if !isWithinBoard(xx, yy) || g.board[yy][xx] == FixedState {
				return true
			}
		}
	}
	return false
}

// removeFullLines checks the board for any full lines and removes them
func (g *Game) removeFullLines() {
	for y := 0; y < gameGridHeight; y++ {
		full := true
		for x := 0; x < gameGridWidth; x++ {
			if g.board[y][x] != FixedState {
				full = false
				break
			}
		}
		if full {
			// Remove the line and shift lines above down
			for yy := y; yy > 0; yy-- {
				for xx := 0; xx < gameGridWidth; xx++ {
					g.board[yy][xx] = g.board[yy-1][xx]
				}
			}
			// Clear the top line
			for x := 0; x < gameGridWidth; x++ {
				g.board[0][x] = EmptyState
			}
		}
	}
}

// moveLeft attempts to move the current piece to the left
func (g *Game) moveLeft() {
	g.lock.Lock()
	defer g.lock.Unlock()
	d := findFirstNonEmptyColumn(shapes[g.piece.Shape.shape][g.piece.Shape.pos])
	if g.piece.Position.x+d-1 >= 0 && !g.hasCollision(g.piece.Position.x-1, g.piece.Position.y, g.piece.Shape.pos) {
		g.piece.Position.x--
	}
	g.tick()
}

// moveRight attempts to move the current piece to the right
func (g *Game) moveRight() {
	g.lock.Lock()
	defer g.lock.Unlock()
	width := getShapeDimensions(shapes[g.piece.Shape.shape][g.piece.Shape.pos])
	x := g.piece.Position.x + findFirstNonEmptyColumn(shapes[g.piece.Shape.shape][g.piece.Shape.pos])
	if x+width < gameGridWidth && !g.hasCollision(g.piece.Position.x+1, g.piece.Position.y, g.piece.Shape.pos) {
		g.piece.Position.x++
	}
	g.tick()
}

// moveDown attempts to move the current piece downwards.
// If the piece cannot move any further, it is set to fixed and a new piece is created
func (g *Game) moveDown() {
	g.lock.Lock()
	defer g.lock.Unlock()
	if g.hasCollision(g.piece.Position.x, g.piece.Position.y+1, g.piece.Shape.pos) {
		if g.piece.Position.y == -getTopOffset(shapes[g.piece.Shape.shape][g.piece.Shape.pos]) {
			g.cancel()
			return
		}
		g.piece.Shape.state = FixedState
	} else {
		g.piece.Shape.state = FallingState
		g.piece.Position.y++
	}
	if g.piece.Shape.state == FixedState {
		g.tick()
		g.newPiece()
	}
	g.tick()
}

// rotate attempts to rotate the current piece.
// If rotation would result in a collision, the piece is left in its current orientation
func (g *Game) rotate() {
	g.lock.Lock()
	defer g.lock.Unlock()
	startY := g.piece.Position.y + getTopOffset(shapes[g.piece.Shape.shape][g.piece.Shape.pos])
	startX := g.piece.Position.x + findFirstNonEmptyColumn(shapes[g.piece.Shape.shape][g.piece.Shape.pos])
	// move shape to the next pos
	g.piece.Shape.pos = (g.piece.Shape.pos + 1) % 4
	// find the next coords for the new shape
	g.piece.Position.x = startX - findFirstNonEmptyColumn(shapes[g.piece.Shape.shape][g.piece.Shape.pos])
	g.piece.Position.y = startY - getTopOffset(shapes[g.piece.Shape.shape][g.piece.Shape.pos])

	for g.piece.Position.x+findFirstNonEmptyColumn(shapes[g.piece.Shape.shape][g.piece.Shape.pos]) >= 0 &&
		g.hasCollision(g.piece.Position.x, g.piece.Position.y, g.piece.Shape.pos) {
		g.piece.Position.x--
	}

	if g.hasCollision(g.piece.Position.x, g.piece.Position.y, g.piece.Shape.pos) {
		g.piece.Shape.pos = g.piece.PreviousShape.pos
		g.piece.Position.x = g.piece.Position.oldX
		g.piece.Position.y = g.piece.Position.oldY
	} else {
		g.moveShape()
	}
	g.tick()
}

// tick updates the game screen to reflect the current state of the game
func (g *Game) tick() {
	g.screen.Clear()
	g.drawMenu()
	g.moveShape()
	g.removeFullLines()
	g.drawBoard()
	g.screen.Show()
}

// runGameLoop runs the main game loop. It listens for user input,
// updates the game state, and handles system signals for graceful shutdown
func (g *Game) runGameLoop(signalChan chan os.Signal) {
	g.tick()
	events := make(chan tcell.Event)

	go func() {
		for {
			events <- g.screen.PollEvent()
		}
	}()

	var wg sync.WaitGroup
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	go func() {
		// This goroutine will listen for a system signal and cancel the context
		// when either SIGINT (Ctrl+C) or SIGTERM is received.
		sig := <-signalChan
		log.Printf("Received signal: %v. Shutting down...", sig)
		g.cancel()
	}()

	wg.Add(1)

	go func() {
		for {
			select {
			case ev := <-events:
				switch event := ev.(type) {
				case *tcell.EventKey:
					g.handleKey(event)
				case *tcell.EventResize:
					g.screen.Sync()
				}
			}
		}
	}()

	for {
		select {
		case <-g.ctx.Done():
			// This will happen when the context is cancelled, i.e., the system signal is received.
			log.Println("Shutting down gracefully...")
			wg.Done()
			os.Exit(0)
		case <-ticker.C:
			g.moveDown()
		}
	}
	wg.Wait()
}

// initializeScreen creates and initializes a new tcell screen.
// It returns an error if the creation or initialization fails
func initializeScreen() (tcell.Screen, error) {
	screen, err := tcell.NewScreen()
	if err != nil {
		log.Fatalln("failed to create new screen:", err)
	}

	if err := screen.Init(); err != nil {
		log.Fatalln("failed to initialize screen:", err)
	}

	return screen, err
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	screen, err := initializeScreen()
	if err != nil {
		os.Exit(1)
	}
	defer screen.Fini()

	screen.EnableMouse()
	screen.Clear()

	g := newGame(screen, ctx, cancel)
	g.runGameLoop(signalChan)
}
