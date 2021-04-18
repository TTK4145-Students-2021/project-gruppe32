package main

import (
	"fmt"

	"os"
	"strconv"

	"./Network/network/bcast"
	"./Network/network/peers"
	"./Requests"
	"./UtilitiesTypes"
	"./elevio"
	"./fsm"
	"./sync"
)

var myElevator UtilitiesTypes.Elevator

func main() {
	sync.LastIncomingMessage.MsgID = 0
	sync.LastIncomingMessage.LocalID = 0

	id := os.Args[1]

	numFloors := 4
	numButtons := 3
	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)
	msgChan := UtilitiesTypes.MsgChan{
		SendChan: make(chan UtilitiesTypes.Msg),
		RecChan:  make(chan UtilitiesTypes.Msg),
	}

	//myElevCh := make(chan UtilitiesTypes.Elevator)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	elevio.Init(fmt.Sprintf("localhost:%s", id), numFloors)
	myElevator.State = fsm.IDLE
	myElevator.ID, _ = strconv.Atoi(id)
	Requests.ClearAllLights(numFloors, numButtons)
	elevio.SetDoorOpenLamp(false)
	fmt.Println("satt init", myElevator.State)

	//sync.Test()
	//go sync.Sync(msgChan, &myElevator)
	peerUpdateCh := make(chan peers.PeerUpdate)
	// We can disable/enable the transmitter after it has been started.
	// This could be used to signal that we are somehow "unavailable".
	peerTxEnable := make(chan bool)

	go peers.Transmitter(10652, id, peerTxEnable)

	go bcast.Transmitter(12569, msgChan.SendChan)

	go bcast.Receiver(12569, msgChan.RecChan)

	go peers.Receiver(10652, peerUpdateCh)

	/*go func() {
		fmt.Println("Started")
		for {
			select {
			case p := <-peerUpdateCh:
				fmt.Printf("Peer update:\n")
				fmt.Printf("  Peers:    %q\n", p.Peers)
				fmt.Printf("  New:      %q\n", p.New)
				fmt.Printf("  Lost:     %q\n", p.Lost)

			}
		}
	}()*/

	//fsm.OnInitBetweenFloors(&myElevator)

	//go fsm.DoorState(&myElevator)
	go sync.SendMessage(msgChan, myElevator)
	go fsm.FSM(msgChan, drv_buttons, drv_floors, &myElevator, peerTxEnable)
	go sync.UpdateOnlineIds(peerUpdateCh, myElevator)

	select {}
}
