package gol

import (
	//"flag"
	"fmt"
	"net/rpc"
	"strconv"
	"sync"
	"time"
	"uk.ac.bris.cs/gameoflife/stubs"
	"uk.ac.bris.cs/gameoflife/util"
)

type distributorChannels struct {
	events     chan<- Event
	ioCommand  chan<- ioCommand
	ioIdle     <-chan bool
	ioFilename chan<- string
	ioOutput   chan<- uint8
	ioInput    <-chan uint8
}

//moved calcAlive to server.go

func makeCall(client *rpc.Client, world [][]byte, params Params) [][]byte {
	params = Params(stubs.Params{Turns: params.Turns, Threads: params.Threads, ImageWidth: params.ImageWidth, ImageHeight: params.ImageHeight})
	request := stubs.Request{World: world, P: stubs.Params(params)}
	response := new(stubs.Response)
	client.Call(stubs.Processsor, request, response)
	return response.World
}

func getalivecall(client *rpc.Client, world [][]byte, params Params) []util.Cell {
	params = Params(stubs.Params{Turns: params.Turns, Threads: params.Threads, ImageWidth: params.ImageWidth, ImageHeight: params.ImageHeight})
	response := new(stubs.AliveResp)
	request := stubs.Request{World: world, P: stubs.Params(params)}
	client.Call(stubs.GetAlive, request, response)
	return response.Alive_Cells
}
func cellsflipped(client *rpc.Client, world [][]byte, params Params, newworld [][]byte) []util.Cell {
	params = Params(stubs.Params{Turns: params.Turns, Threads: params.Threads, ImageWidth: params.ImageWidth, ImageHeight: params.ImageHeight})
	response := new(stubs.AliveResp)
	request := stubs.Request2{World: world, P: stubs.Params(params), NewWorld: newworld}
	client.Call(stubs.GetCellsFlipped, request, response)
	return response.Alive_Cells
}

func cancelserver(client *rpc.Client) bool {
	client.Call(stubs.CancelServer, stubs.EmptyReq{}, stubs.EmptyReq{})
	return true
}

func broker(client *rpc.Client, world [][]byte, params Params, newworld [][]byte) {
	params = Params(stubs.Params{Turns: params.Turns, Threads: params.Threads, ImageWidth: params.ImageWidth, ImageHeight: params.ImageHeight})
	request := stubs.Request2{World: world, P: stubs.Params(params), NewWorld: newworld}
	response := new(stubs.Response)
	client.Call(stubs.CalculateAlive, request, response)
}

/// basically make new method for rpc call and then put that in th go func before the resval thing.
//this is my implementation of it at least for the moment
//func cancelserver(client rpc.Client) *stubs.E

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p Params, c distributorChannels, keyPresses <-chan rune) {
	//, keyPresses <-chan rune
	//string conversion for the filename

	height := strconv.Itoa(p.ImageHeight)
	width := strconv.Itoa(p.ImageWidth)

	FileName := width + "x" + height
	c.ioCommand <- ioInput

	c.ioFilename <- FileName

	// TODO: Create a 2D slice to store the world.

	world := make([][]byte, p.ImageHeight)
	for i := range world {
		world[i] = make([]byte, p.ImageWidth)
	}

	//updating it with the bytes sent from io.go
	ogworld := world //to use with the cellflipped function
	var lock sync.Mutex

	for row := 0; row < p.ImageHeight; row++ {
		for col := 0; col < p.ImageWidth; col++ {
			world[row][col] = <-c.ioInput //not sure if it works
			if world[row][col] == 255 {
				cell := util.Cell{X: row, Y: col}
				c.events <- CellFlipped{CompletedTurns: 0, Cell: cell}
			}
		}

	}

	// TODO: Execute all turns of the Game of Life.
	tk := time.NewTicker(2 * time.Second)

	var turn = 0
	var ran = 0

	go func() {
		for {
			select {
			case <-tk.C:
				//lock.Lock()
				c.events <- AliveCellsCount{turn, len(getalivecall(client, world, p))}
				//lock.Unlock()
			case x := <-keyPresses:
				if x == 'p' {
					fmt.Println(turn)
					fmt.Println("Paused!")
					lock.Lock()
					for {
						if <-keyPresses == 'p' {
							lock.Unlock()
							fmt.Println("Continuing..")
							break
						}
					}
				} else if x == 's' {
					c.ioCommand <- ioOutput

					fturn := strconv.Itoa(turn)

					FileName = width + "x" + height + "x" + fturn

					c.ioFilename <- FileName

					for row := 0; row < p.ImageHeight; row++ {
						for col := 0; col < p.ImageWidth; col++ {
							c.ioOutput <- world[row][col]
						}
					}

					c.events <- ImageOutputComplete{CompletedTurns: turn, Filename: FileName}
				} else if x == 'q' {
					c.ioCommand <- ioOutput
					fturn := strconv.Itoa(turn)
					FileName = width + "x" + height + "x" + fturn
					c.ioFilename <- FileName
					for row := 0; row < p.ImageHeight; row++ {
						for col := 0; col < p.ImageWidth; col++ {
							c.ioOutput <- world[row][col]
						}
					}
					c.events <- ImageOutputComplete{CompletedTurns: turn, Filename: FileName}
					c.events <- FinalTurnComplete{CompletedTurns: turn, Alive: getalivecall(client, world, p)}
					fmt.Println("Terminated.")

				} else if x == 'k' { //should cancel server from within itself
					ran = 1
					//cancelserver(client)
				}

			}
		}

	}()

	for turn < p.Turns {
		//if stop{//if paused stops at this turn till stop becomes false
		//for stop{
		lock.Lock()
		lock.Unlock()
		ogworld = world
		world = makeCall(client, world, p) //getting new state
		cd := cellsflipped(client, world, p, ogworld)
		for _, s := range cd {
			c.events <- CellFlipped{CompletedTurns: turn, Cell: s}
		}
		c.events <- TurnComplete{turn}
		turn++
		if ran == 1 {
			cancelserver(client)
			c.ioCommand <- ioOutput
			fturn := strconv.Itoa(turn)
			FileName = width + "x" + height + "x" + fturn
			c.ioFilename <- FileName
			for row := 0; row < p.ImageHeight; row++ {
				for col := 0; col < p.ImageWidth; col++ {
					c.ioOutput <- world[row][col]
				}
			}
			c.events <- ImageOutputComplete{CompletedTurns: turn, Filename: FileName}
			c.events <- FinalTurnComplete{CompletedTurns: turn, Alive: getalivecall(client, world, p)}
			fmt.Println("Terminated.")
		}
	}
	tk.Stop()

	//stage 3
	c.ioCommand <- ioOutput
	turnstr := strconv.Itoa(p.Turns)
	name := height + "x" + width + "x" + turnstr
	c.ioFilename <- name
	for row := 0; row < p.ImageHeight; row++ {
		for col := 0; col < p.ImageWidth; col++ {
			c.ioOutput <- world[row][col]
		}
	}
	c.events <- ImageOutputComplete{p.Turns, name}

	// report how many turns are over and how many alivers are remaining after each turn
	// and send that to a channel which goes into the FinalTurnComplete

	// TODO: Report the final st	resval := makeCall(*client, world, p)

	alivers := getalivecall(client, world, p)

	final := FinalTurnComplete{CompletedTurns: p.Turns, Alive: alivers}
	c.events <- final

	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle

	c.events <- StateChange{turn, Quitting}

	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)
}
