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

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	elevio.Init(fmt.Sprintf("localhost:%s", id), numFloors)
	myElevator.State = fsm.IDLE
	myElevator.ID, _ = strconv.Atoi(id)
	Requests.ClearAllLights(numFloors, numButtons)

	//sync.Test()
	//go sync.Sync(msgChan, &myElevator)

	go bcast.Transmitter(12569, msgChan.SendChan)
	go bcast.Receiver(12569, msgChan.RecChan)
	go sync.SendMessage(msgChan)

	peerUpdateCh := make(chan peers.PeerUpdate)
	// We can disable/enable the transmitter after it has been started.
	// This could be used to signal that we are somehow "unavailable".
	peerTxEnable := make(chan bool)
	go peers.Transmitter(10652, id, peerTxEnable)
	go peers.Receiver(10652, peerUpdateCh)

	go func() {
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
	}()

	go fsm.DoorState(&myElevator)

	for {
		select {
		case a := <-drv_buttons:
			if a.Button == elevio.BT_Cab {
				fsm.OnRequestButtonPress(&myElevator, a.Floor, a.Button)
				sync.AddElevToMsgQueue(myElevator)
			} else {
				sync.AddHallOrderToMsgQueue(myElevator, a.Floor, a.Button)
			}
		case a := <-drv_floors:
			fsm.OnFloorArrival(&myElevator, a)
			sync.AddElevToMsgQueue(myElevator)

		case incomingMsg := <-msgChan.RecChan:
			sync.Run(incomingMsg, myElevator, msgChan)
			if sync.ShouldITake(incomingMsg, myElevator) {
				//fmt.Println(sync.OnlineElevators[1].Orders[1][1].Status)
				//fmt.Println(sync.OnlineElevators[1].ID)
				myElevator.Orders[incomingMsg.Order.Floor][incomingMsg.Order.ButtonType].Status = UtilitiesTypes.Active
				sync.AddElevToMsgQueue(myElevator)
				fsm.OnRequestButtonPress(&myElevator, incomingMsg.Order.Floor, elevio.ButtonType(incomingMsg.Order.ButtonType))
				sync.AddElevToMsgQueue(myElevator)
			}

		}
	}

	//FsmFunction(drv_buttons, drv_floors, drv_obstr, drv_stop)
	//var noOrder UtilitiesTypes.Order
	//noOrder = UtilitiesTypes.Order{Floor: -1, ButtonType: -1, Status: -1, Finished: false}
	//myElev := UtilitiesTypes.Elevator{Dir: 0, Floor: -1, State: 0, Orders: [numFloors][numButtons]noOrder}
	//f:= 0
	//currentOrder := elevio.ButtonEvent{Floor: -1, Button: -1}

	/*
			for {
				select {
				case a := <-drv_buttons:
					fsm.OnRequestButtonPress(a.Floor, a.Button)

				case a := <-drv_floors:
					fsm.OnFloorArrival(a)

				case a := <-drv_stop:
					fmt.Printf("%+v\n", a)
					for f := 0; f < numFloors; f++ {
						for b := elevio.ButtonType(0); b < 3; b++ {
							elevio.SetButtonLamp(b, f, false)
						}
					}
				}
			}

		}

		/*var d elevio.MotorDirection = elevio.MD_Up
			//elevio.SetMotorDirection(d)

			drv_buttons := make(chan elevio.ButtonEvent)
			drv_floors := make(chan int)
			drv_obstr := make(chan bool)
			drv_stop := make(chan bool)

			go elevio.PollButtons(drv_buttons)
			go elevio.PollFloorSensor(drv_floors)
			go elevio.PollObstructionSwitch(drv_obstr)
			go elevio.PollStopButton(drv_stop)
			go orderhandler.UpdateLights(drv_buttons)

			for {
				select {
				//case a := <-drv_buttons:
					//orderhandler.UpdateLights(a)
					//fmt.Printf("%+v\n", a)
					//elevio.SetButtonLamp(a.Button, a.Floor, true)

				case a := <-drv_floors:
					fmt.Printf("%+v\n", a)
					if a == numFloors-1 {
						d = elevio.MD_Down
					} else if a == 0 {
						d = elevio.MD_Up
					}
					elevio.SetMotorDirection(d)

				case a := <-drv_obstr:
					fmt.Printf("%+v\n", a)
					if a {
						elevio.SetMotorDirection(elevio.MD_Stop)
					} else {
						elevio.SetMotorDirection(d)
					}

				case a := <-drv_stop:
					fmt.Printf("%+v\n", a)
					for f := 0; f < numFloors; f++ {
						for b := elevio.ButtonType(0); b < 3; b++ {
							elevio.SetButtonLamp(b, f, false)
						}
					}
				}
			}*/
}
