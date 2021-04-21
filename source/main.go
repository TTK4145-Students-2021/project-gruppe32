package main

import (
	"fmt"
	"os"
	"strconv"

	eio "./ElevIO"
	"./FSM"
	"./Network/network/bcast"
	"./Network/network/peers"
	Req "./Requests"
	"./Sync"
	UT "./UtilitiesTypes"
)

var myElevator UT.Elevator

func main() {
	Sync.LastIncomingMessage.MsgID = 0
	Sync.LastIncomingMessage.LocalID = 0

	id := os.Args[1]

	const NumFloors = UT.NumFloors
	const NumButtons = UT.NumButtons

	drv_buttons := make(chan eio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)

	msgChan := UT.MsgChan{
		SendChan: make(chan UT.Msg),
		RecChan:  make(chan UT.Msg),
	}

	go eio.PollButtons(drv_buttons)
	go eio.PollFloorSensor(drv_floors)
	go eio.PollObstructionSwitch(drv_obstr)
	go eio.PollStopButton(drv_stop)

	eio.Init(fmt.Sprintf("localhost:%s", id), NumFloors)
	myElevator.State = FSM.IDLE
	myElevator.ID, _ = strconv.Atoi(id)
	Req.ClearAllLights(NumFloors, NumButtons)
	eio.SetDoorOpenLamp(false)

	peerUpdateCh := make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool)

	go peers.Transmitter(10652, id, peerTxEnable)
	go bcast.Transmitter(12569, msgChan.SendChan)
	go bcast.Receiver(12569, msgChan.RecChan)
	go peers.Receiver(10652, peerUpdateCh)

	go Sync.SendMessage(msgChan, myElevator)
	go FSM.FSM(msgChan, drv_buttons, drv_floors, &myElevator, peerTxEnable, drv_obstr)
	go Sync.UpdateFromPeers(peerUpdateCh, myElevator)

	select {}
}
