package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net"
	"net/rpc"
	"os"
	//"strconv"
	"time"
	"uk.ac.bris.cs/gameoflife/stubs"
	"uk.ac.bris.cs/gameoflife/util"
)


func calcN(client *rpc.Client, world [][] byte, params Params, newworld [][]byte){
	params = Params(stubs.Params{Turns: params.Turns, Threads: params.Threads, ImageWidth: params.ImageWidth, ImageHeight: params.ImageHeight})
	request := stubs.Request2{World: world,P: stubs.Params(params), NewWorld: newworld}
	response := new(stubs.Response)
	client.Call(stubs.CalculateNext, request, response)

}
func calcA(client *rpc.Client, world [][] byte, params Params, newworld [][]byte){
	params = Params(stubs.Params{Turns: params.Turns, Threads: params.Threads, ImageWidth: params.ImageWidth, ImageHeight: params.ImageHeight})
	request := stubs.Request2{World: world,P: stubs.Params(params), NewWorld: newworld}
	response := new(stubs.Response)
	client.Call(stubs.CalculateAlive, request, response)
}
}

type GameofLifeOperations struct{}

//var alivecells int

func (s *GameofLifeOperations) Process(req stubs.Request, res *stubs.Response) (err error) {
	// take the parameters from the req util thingy
	var world = req.World
	world = calculateNextState(req.P, world)

	// send the next turn stRequestuff thru to the response struct
	res.World = world

	return
}
func (s *GameofLifeOperations) GetAlivers(req stubs.Request, res *stubs.AliveResp) (err error) {
	res.Alive_Cells = calculateAliveCells(req.P, req.World)
	return
}
func (s *GameofLifeOperations) GetCellsFlipped(req stubs.Request2, res *stubs.AliveResp) (err error) {
	newWorldData := req.NewWorld
	world := req.World
	returnable := make([]util.Cell, 0)
	for row := 0; row < req.P.ImageHeight; row++ {
		for col := 0; col < req.P.ImageWidth; col++ {
			if newWorldData[row][col] != world[row][col] {
				cell := util.Cell{X: row, Y: col}
				returnable = append(returnable, cell)
			}
		}
	}
	//c.events <- TurnComplete{CompletedTurns: turn}
	res.Alive_Cells = returnable
	return
}
func (s *GameofLifeOperations) CancelServer(req stubs.EmptyReq, res *stubs.ServerCancelled) (err error) {
	os.Exit(0)
	return
}



func main() {

 	//from distributor
	server := "127.0.0.1:8030"
	flag.Parse()
	client, b := rpc.Dial("tcp", server)
	if b != nil {
		fmt.Println(b)
	}
	defer client.Close()

	//from server
	pAddr := flag.String("port", "8030", "Port to listen on")
	flag.Parse()
	rand.Seed(time.Now().UnixNano())
	rpc.Register(&GameofLifeOperations{})
	listener, _ := net.Listen("tcp", ":"+*pAddr)
	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {

		}
	}(listener)
	rpc.Accept(listener)
}
var serworld [][]byte

var disworld [][]byte


func (s *GameofLifeOperations) broker (p stubs.Params, c distributorChannels)

	turn:=0
	for turn < p.Turns {
		//if stop{//if paused stops at this turn till stop becomes false
		//for stop{
		//lock.Lock()
		//lock.Unlock()
		disworld = world
		world = makeCall(client, world, p) //getting new state
		cd := cellsflipped(client, world, p, ogworld)
		for _, s := range cd {
			c.events <- CellFlipped{CompletedTurns: turn, Cell: s}
		}
		c.events <- TurnComplete{turn}
		turn++

		}
	}
}