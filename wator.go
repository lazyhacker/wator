// wator is an implementation of the Wa-Tor simulation in Go.
//
// The rules are the simulation are:
//
//   - The world is a toroidal (donut-shaped) sea planet consisting of fish and
//     sharks.
//   - Fish feed on ubiuitous plankton and the sharks feed on the fish.
//   - During each cycle fish move randomly to an unoccupied adjacent square.
//   - After a number of cycles will spawn a new fish.
//   - Sharks will move to an adjacent square if there is a fish and eats the
//     fish otherwise it will move to an random adjacent unoccupied square.
//   - Sharks must eat a fish within a number of cycles or it will die.
//   - At a certain age a shark will spawn a new shark.

package main // import "lazyhacker.dev/wator"

import (
	"flag"
	"fmt"
	"image/color"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
)

// Define the directions fish/sharks can go.
const (
	NORTH = iota
	SOUTH
	EAST
	WEST
)

type coordinate struct {
	x, y int
}

var (
	nFish   = flag.Int("fish", 10, "Initial # of fish.")
	nSharks = flag.Int("sharks", 1000, "Initial # of sharks.")
	fBreed  = flag.Int("fbreed", 20, "# of cycles for fish to reproduce (default: 20)")
	sBreed  = flag.Int("sbreed", 30, "# of cycles for shark to reproduce (default: 40)")
	starve  = flag.Int("starve", 30, "# of cycles shark can go with feeding before dying (default: 15)")
	wwidth  = flag.Int("width", 320, "Width of the world (East - West).")
	wheight = flag.Int("height", 240, "Height of the world (North-South).")
)

var tick = 0
var wm [][]*creature

// Types of creatures in Wator.
const (
	FISH = iota
	SHARK
)

var (
	fishcolor  = color.RGBA{255, 255, 0, 255} // YELLOW
	sharkcolor = color.RGBA{255, 0, 0, 255}   // RED
)

type creature struct {
	age, health, species int
	asset                color.RGBA
	chronon              int
}

// Chronon determines what happens with the world at each turn.
func Chronon(c int) {

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	var xcoord, ycoord int

	for y := 0; y < *wheight; y++ {
		for x := 0; x < *wwidth; x++ {

			north, south, east, west := adjacent(x, y)

			if wm[x][y] == nil {
				continue
			}

			if wm[x][y].chronon == c {
				continue
			}
			wm[x][y].age += 1
			d := r.Intn(3)
			switch wm[x][y].species {
			case FISH:
				foundspace := false
				for i := 0; i < 4; i++ {
					d += i

					switch d % 4 {
					case NORTH:
						xcoord = north.x
						ycoord = north.y
					case SOUTH:
						xcoord = south.x
						ycoord = south.y
					case EAST:
						xcoord = east.x
						ycoord = east.y
					case WEST:
						xcoord = west.x
						ycoord = west.y
					}

					// Found an open square.
					if wm[xcoord][ycoord] == nil {
						foundspace = true
						wm[xcoord][ycoord] = wm[x][y]

						//log.Printf("Moving fish from (%d, %d)  to (%d, %d)\n", x, y, xcoord, ycoord)

						// If not a baby and of spawning age.
						if wm[x][y].age != 0 && wm[x][y].age%*fBreed == 0 {
							// spawn a new fish in its place
							wm[x][y] = &creature{
								age:     0,
								species: FISH,
								asset:   fishcolor,
								chronon: c,
							}
						} else {
							wm[x][y] = nil
						}
						break
					}
				}
				if !foundspace {
					wm[x][y] = nil
				}
			case SHARK:
				//log.Printf("Shark at (%d, %d)\n", x, y)

				foundfish := false
				wm[x][y].health -= 1

				if wm[x][y].health == 0 {
					wm[x][y] = nil
					break
				}
				wm[x][y].chronon = c

				for i := 0; i < 4; i++ {
					d += i
					switch d % 4 {
					case NORTH:
						xcoord = north.x
						ycoord = north.y
					case SOUTH:
						xcoord = south.x
						ycoord = south.y
					case EAST:
						xcoord = east.x
						ycoord = east.y
					case WEST:
						xcoord = west.x
						ycoord = west.y
					}

					if wm[xcoord][ycoord] == nil {
						break
					}

					// Found a fish in adjacent square so eat it and move there.
					if wm[xcoord][ycoord].species == FISH {
						foundfish = true
						wm[xcoord][ycoord] = wm[x][y]
						wm[xcoord][ycoord].health = *starve
						break
					}
				}

				// If no fish, pick adjacent square and move there.
				if !foundfish {
					for i := 0; i < 4; i++ {
						d += i
						switch d % 4 {
						case NORTH:
							xcoord = north.x
							ycoord = north.y
						case SOUTH:
							xcoord = south.x
							ycoord = south.y
						case EAST:
							xcoord = east.x
							ycoord = east.y
						case WEST:
							xcoord = west.x
							ycoord = west.y
						}

						if wm[xcoord][ycoord] == nil {
							wm[xcoord][ycoord] = wm[x][y]
							wm[xcoord][ycoord].chronon = c
							wm[x][y] = nil

							// Spawn a new shark in the old spot.
							if wm[xcoord][ycoord].age != 0 && wm[xcoord][ycoord].age%*sBreed == 0 {
								wm[x][y] = &creature{
									age:     0,
									health:  *starve,
									species: SHARK,
									asset:   sharkcolor,
									chronon: c,
								}
							}
							break
						}
					}
				}
			}
		}
	}
}

// adjacent returns the adjecent squares in the order of
// North, South, East, West.
func adjacent(x, y int) (coordinate, coordinate, coordinate, coordinate) {

	var n, s, e, w coordinate
	if y == 0 {
		n.y = *wheight - 1
	} else {
		n.y = y - 1
	}
	n.x = x
	if y == *wheight-1 {
		s.y = 0
	} else {
		s.y = y + 1
	}
	s.x = x
	if x == *wwidth-1 {
		e.x = 0
	} else {
		e.x = x + 1
	}
	e.y = y
	if x == 0 {
		w.x = *wwidth - 1
	} else {
		w.x = x - 1
	}
	w.y = y

	return n, s, e, w

}
func initWator() [][]*creature {

	// Set up the world map as a 2-D Slice
	var wm = make([][]*creature, *wwidth)
	for i := range wm {
		wm[i] = make([]*creature, *wheight)
	}
	pop := 0

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	//r := rand.New(rand.NewSource(10))
	for i := 0; i < *nFish; i++ {
		for {
			if pop == *wwidth**wheight {
				break
			}
			x := r.Intn(*wwidth - 1)
			y := r.Intn(*wheight - 1)

			if wm[x][y] == nil {
				wm[x][y] = &creature{
					age:     0,
					species: FISH,
					asset:   fishcolor,
				}
				pop++
				break
			}
		}
	}

	for i := 0; i < *nSharks; i++ {
		for {
			if pop == *wwidth**wheight {
				break
			}
			x := r.Intn(*wwidth - 1)
			y := r.Intn(*wheight - 1)

			if wm[x][y] == nil {
				wm[x][y] = &creature{
					age:     0,
					species: SHARK,
					health:  *starve,
					asset:   sharkcolor,
				}
				pop++
				break
			}
		}
	}

	return wm
}

func debug() {
	for y := 0; y < *wheight; y++ {
		for x := 0; x < *wwidth; x++ {
			if wm[x][y] == nil {
				fmt.Print(" ")
			} else {
				switch wm[x][y].species {
				case FISH:
					fmt.Print("F")
				case SHARK:
					fmt.Print("S")
				}
			}
		}

		fmt.Println()
	}
}

func update(screen *ebiten.Image) error {

	tick++
	Chronon(tick)

	if ebiten.IsDrawingSkipped() {
		return nil
	}

	screen.Fill(color.RGBA{255, 255, 255, 255})
	render(screen)
	ebitenutil.DebugPrint(screen, strconv.Itoa(tick))
	return nil

}

func render(screen *ebiten.Image) {
	for x := 0; x < *wwidth; x++ {
		for y := 0; y < *wheight; y++ {
			if wm[x][y] != nil {
				screen.Set(x, y, wm[x][y].asset)
			} else {
				screen.Set(x, y, color.RGBA{0, 0, 0, 255})
			}
		}
	}
}

func main() {

	flag.Parse()

	if *nFish+*nSharks > *wwidth**wheight {
		log.Fatal("Not enough space for Fish and Shark!")
	}

	wm = initWator()

	if err := ebiten.Run(update, *wwidth, *wheight, 2, "Wator"); err != nil {
		log.Fatal(err)
	}
}
