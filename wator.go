// This is an implementation of Wator in Go.
package main // import "lazyhacker.dev/wator"

import (
	"container/list"
	"flag"
	"math/rand"
	"time"

	termbox "github.com/nsf/termbox-go"
)

// Define the directions fish/sharks can go.
const (
	NORTH = iota
	SOUTH
	EAST
	WEST
)

// Types of creatures in Wator.
const (
	PLANKTON = iota
	FISH
	SHARK
)

// Constant value for various fish/shark settings.
const (
	FISHSPAWN   = 20 // rounds for fish to give birth
	SHARKSPAWN  = 40 // rounds for shark to give birth
	SHARKHEALTH = 15 // rounds shark survives without eating
)

var (
	initFish    = *flag.Int("fish", 50, "Initial # of fish.")
	initShark   = *flag.Int("sharks", 20, "Initial # of sharks.")
	worldWidth  = *flag.Int("width", 20, "Width of the world (East - West).")
	worldHeight = *flag.Int("height", 20, "Height of the world (North-South).")
)

// Fish represents a fish that eats plankton and give birth after X number of turns.
type Fish struct {
	spawn int // counter to birthing another fish
	x, y  int // position on the map
}

// Shark represents a shark.
type Shark struct {
	spawn, health int
	x, y          int
}

// MapNode represent each position of a world map.
type MapNode struct {
	ctype    int         // creature type
	creature interface{} // pointer to the fish or shark
}

// WorldMap represents a map of Wator.
type WorldMap [][]MapNode

// SetMapNode will write to values of a MapNode.
func SetMapNode(wm WorldMap, x int, y int, ct int, c interface{}) {
	wm[x][y].ctype = ct
	wm[x][y].creature = c
}

// DrawMap will draw the current state of the world map.
func DrawMap(m WorldMap) {
	for i := 0; i < worldWidth; i++ {
		for j := 0; j < worldHeight; j++ {
			switch m[i][j].ctype {
			case FISH:
				termbox.SetCell(i, j, 'F', termbox.ColorYellow, termbox.ColorBlue)
			case SHARK:
				termbox.SetCell(i, j, 'S', termbox.ColorRed, termbox.ColorBlue)
			default:
				termbox.SetCell(i, j, 'P', termbox.ColorBlue, termbox.ColorBlue)
			}
		}
	}
}

// GetDirection will return a new random direction from origin.
func GetDirection(x int, y int) (int, int, int) {
	d := rand.Intn(4)

	nx := x
	ny := y
	switch d {
	case NORTH:
		if ny == 0 {
			ny = worldHeight - 1
		} else {
			ny--
		}
	case SOUTH:
		ny++
		if ny == worldHeight {
			ny = 0
		}
	case EAST:
		nx++
		if nx == worldWidth {
			nx = 0
		}
	case WEST:
		if nx == 0 {
			nx = worldWidth - 1
		} else {
			nx--
		}
	}

	return nx, ny, d
}

// GetTermboxEvents will get the keyboard event and pass it to the event queue.
func GetTermboxEvents(evtQueue chan<- termbox.Event) {
	for {
		evtQueue <- termbox.PollEvent()
	}
}

func main() {

	rand.Seed(time.Now().Unix())
	flag.Parse()

	bklp := false
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	if initFish+initShark > worldWidth*worldHeight {
		panic("Not enough space for Fish and Shark!")
	}

	// Set up the world map as a 2-D Slice
	wm := make(WorldMap, worldHeight)
	for i := range wm {
		wm[i] = make([]MapNode, worldWidth)
	}

	slist := list.New() // list of sharks
	flist := list.New() // list of fish

	// Create the initial set of fish.
	for i := 0; i < initFish; i++ {
		f := new(Fish)
		for {
			x := rand.Intn(worldWidth - 1)
			y := rand.Intn(worldHeight - 1)
			if wm[x][y].ctype == PLANKTON {
				f.x = x
				f.y = y
				break
			}
		}
		wm[f.x][f.y].ctype = FISH
		wm[f.x][f.y].creature = flist.PushBack(f)
	}

	// Create the initial set of sharks.
	for i := 0; i < initShark; i++ {
		s := new(Shark)
		for {
			x := rand.Intn(worldWidth)
			y := rand.Intn(worldHeight)
			if wm[x][y].ctype == PLANKTON {
				wm[x][y].ctype = SHARK
				s.x = x
				s.y = y
				break
			}
		}
		wm[s.x][s.y].creature = slist.PushBack(s)
	}
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	DrawMap(wm)
	termbox.Flush()

	// Game loop
	eq := make(chan termbox.Event) // channel to pass keyboard events
	var dv [4]bool                 // array that tracks direction tried
	for {

		// Listen for keyboard even to signal existing game loop.
		go GetTermboxEvents(eq)
		select {
		case ev := <-eq:
			if ev.Type == termbox.EventKey && ev.Key == termbox.KeyEsc {
				bklp = true
			}
		default:
			bklp = false
		}

		termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
		// Loop through each fish.
		for e := flist.Front(); e != nil; e = e.Next() {
			f := e.Value.(*Fish)
			for i := range dv {
				dv[i] = false
			}
			for { // Loop until an unoccupied direction is found.
				nx, ny, d := GetDirection(f.x, f.y)
				dv[d] = true

				f.spawn++

				if wm[nx][ny].ctype == PLANKTON {
					// Determine if we spawn a new fish.  If so then put in the orig spot.
					if f.spawn == FISHSPAWN && flist.Len()+slist.Len() < worldWidth*worldHeight {
						nf := new(Fish)
						nf.x = f.x
						nf.y = f.y
						nf.spawn = 0
						f.spawn = 0
						wm[nf.x][nf.y].ctype = FISH
						wm[nf.x][nf.y].creature = flist.PushBack(nf)
					} else {
						wm[f.x][f.y].ctype = PLANKTON
						wm[f.x][f.y].creature = nil
					}
					wm[nx][ny].ctype = FISH
					wm[nx][ny].creature = e
					f.x = nx
					f.y = ny
					break
				}

				if dv[0] && dv[1] && dv[2] && dv[3] {
					break // No unoccupied adjacent space.
				}
			}
		}

		// Loop through each shark.
		for e := slist.Front(); e != nil; e = e.Next() {
			s := e.Value.(*Shark)
			for i := range dv {
				dv[i] = false
			}
			for {
				nx, ny, d := GetDirection(s.x, s.y)
				dv[d] = true
				s.spawn++
				s.health++

				if wm[nx][ny].ctype == PLANKTON || wm[nx][ny].ctype == FISH {
					if s.spawn == SHARKSPAWN && flist.Len()+slist.Len() < worldWidth*worldHeight {
						ns := new(Shark)
						ns.x = s.x
						ns.y = s.y
						ns.spawn = 0
						s.spawn = 0
						wm[ns.x][ns.y].ctype = SHARK
						wm[ns.x][ns.y].creature = slist.PushBack(ns)
					} else {
						wm[s.x][s.y].ctype = PLANKTON
						wm[s.x][s.y].creature = nil
					}
					if wm[nx][ny].ctype == FISH {
						flist.Remove(wm[nx][ny].creature.(*list.Element))
						wm[nx][ny].creature = nil
						s.health = 0
					}

					if s.health == SHARKHEALTH {
						wm[nx][ny].ctype = PLANKTON
						wm[nx][ny].creature = nil
						slist.Remove(e)
					} else {
						wm[nx][ny].ctype = SHARK
						wm[nx][ny].creature = e
						s.x = nx
						s.y = ny
					}
					break
				}
				if dv[0] && dv[1] && dv[2] && dv[3] {
					break
				}
			}
		}
		if bklp {
			break
		}
		DrawMap(wm)
		termbox.Flush()
		time.Sleep(time.Duration(1) * time.Second)
	}
}
