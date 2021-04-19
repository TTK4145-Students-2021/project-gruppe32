package main

import (
	"fmt"
	"os"
	"strconv"

	"./Network/network/bcast"
	"./Network/network/peers"
	Req"./Requests"
	UT"./UtilitiesTypes"
	eio"./elevio"
	"./fsm"
	"./sync"
)

var myElevator UT.Elevator

func main() {
	sync.LastIncomingMessage.MsgID = 0
	sync.LastIncomingMessage.LocalID = 0

	id := os.Args[1]

	const NumFloors  = UT.NumFloors
	const NumButtons = UT.NumButtons

	drv_buttons := make(chan eio.ButtonEvent)
	drv_floors  := make(chan int)
	drv_obstr   := make(chan bool)
	drv_stop    := make(chan bool) 

	msgChan := UT.MsgChan{
		SendChan: make(chan UT.Msg),
		RecChan:  make(chan UT.Msg),
	}

	go eio.PollButtons(drv_buttons)
	go eio.PollFloorSensor(drv_floors)
	go eio.PollObstructionSwitch(drv_obstr)
	go eio.PollStopButton(drv_stop)

	eio.Init(fmt.Sprintf("localhost:%s", id), NumFloors)
	myElevator.State = fsm.IDLE
	myElevator.ID, _ = strconv.Atoi(id)
	Req.ClearAllLights(NumFloors, NumButtons)
	eio.SetDoorOpenLamp(false)
	fmt.Println("satt init", myElevator.State)

	//go sync.Sync(msgChan, &myElevator)
	
	peerUpdateCh := make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool)

	go peers.Transmitter(10652, id, peerTxEnable)
	go bcast.Transmitter(12569, msgChan.SendChan)
	go bcast.Receiver(12569, msgChan.RecChan)
	go peers.Receiver(10652, peerUpdateCh)

	//fsm.OnInitBetweenFloors(&myElevator)

	//go fsm.DoorState(&myElevator)
	go sync.SendMessage(msgChan, myElevator)
	go fsm.FSM(msgChan, drv_buttons, drv_floors, &myElevator, peerTxEnable, drv_obstr)
	go sync.UpdateOnlineIds(peerUpdateCh, myElevator)

	select {}
}
