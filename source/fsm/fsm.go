package fsm

import (
	"fmt"

	"time"

	"../Requests"
	"../UtilitiesTypes"
	"../elevio"
	"../sync"
)

const numFloors = 4

const numButtons = 3

type State int

const (
	INIT      State = 0
	IDLE            = 1
	MOVING          = 2
	DOOR            = 3
	UNDEFINED       = 4
)

//var state State

func OnInitBetweenFloors(myElev *UtilitiesTypes.Elevator) {
	elevio.SetMotorDirection(elevio.MD_Down)
	myElev.Dir = elevio.MD_Down
	myElev.State = UtilitiesTypes.MOVING
}

/*
func DoorState(myElev *UtilitiesTypes.Elevator) {
	for {
		if Requests.TimeOut(3, *myElev) {
			fmt.Println("TimerOut")
			OnDoorTimeout(myElev)
		}
	}
}*/

func FSM(msgChan UtilitiesTypes.MsgChan, drv_buttons chan elevio.ButtonEvent, drv_floors chan int, myElev *UtilitiesTypes.Elevator, peerCh chan bool) {
	doorTimeout := time.NewTimer(3 * time.Second)
	engineErrorTimeout := time.NewTimer(5 * time.Second)
	doorTimeout.Stop()
	engineErrorTimeout.Stop()
	for {
		sync.UpdateHallLights()
		select {
		case btn := <-drv_buttons:
			if btn.Button == elevio.BT_Cab {
				fmt.Println("ny cab", myElev.State)
				switch myElev.State {
				case DOOR:
					fmt.Println("door")
					if !(myElev.Floor == btn.Floor) {
						myElev.Orders[btn.Floor][btn.Button].Status = UtilitiesTypes.Active
					} else {
						engineErrorTimeout.Reset(5 * time.Second)
						myElev.Orders[btn.Floor][btn.Button].Status = UtilitiesTypes.Inactive
						elevio.SetDoorOpenLamp(true)
						doorTimeout.Reset(3 * time.Second)
					}
					break

				case MOVING:
					myElev.Orders[btn.Floor][btn.Button].Status = UtilitiesTypes.Active
					break

				case IDLE:
					fmt.Println("før if")
					if myElev.Floor == btn.Floor {
						fmt.Println("etter if")
						elevio.SetDoorOpenLamp(true)
						doorTimeout.Reset(3 * time.Second)
						myElev.Orders[btn.Floor][btn.Button].Status = UtilitiesTypes.Inactive
						myElev.State = UtilitiesTypes.DOOR
					} else {
						engineErrorTimeout.Reset(3 * time.Second)
						myElev.Orders[btn.Floor][btn.Button].Status = UtilitiesTypes.Active
						myElev.Dir = Requests.ChooseDirection(*myElev)
						elevio.SetMotorDirection(myElev.Dir)
						myElev.State = UtilitiesTypes.MOVING
					}
					break
				}
				Requests.SetAllCabLights(*myElev, numFloors, numButtons)
				sync.AddElevToMsgQueue(*myElev)
			} else {
				sync.AddHallOrderToMsgQueue(*myElev, btn.Floor, btn.Button)
			}
		case newFloor := <-drv_floors:
			myElev.Floor = newFloor
			myElev.MotorStop = false
			peerCh <- true
			engineErrorTimeout.Reset(5 * time.Second)

			elevio.SetFloorIndicator(myElev.Floor)

			if Requests.ShouldStop(*myElev) {
				elevio.SetMotorDirection(elevio.MD_Stop)
				elevio.SetDoorOpenLamp(true)
				Requests.ClearAtCurrentFloor(myElev, numFloors, numButtons)
				doorTimeout.Reset(3 * time.Second)
				engineErrorTimeout.Stop()
				Requests.SetAllCabLights(*myElev, numFloors, numButtons)
				myElev.State = UtilitiesTypes.DOOR
			} else if myElev.State == MOVING {
				engineErrorTimeout.Reset(3 * time.Second)
			}

			sync.AddElevToMsgQueue(*myElev)

		case incomingMsg := <-msgChan.RecChan:
			//MsgTimeout := time.NewTimer(20 * time.Millisecond)
			sync.Run(incomingMsg, *myElev, msgChan)
			if sync.ShouldITake(incomingMsg, *myElev) {
				//fmt.Println(sync.OnlineElevators[1].Orders[1][1].Status)
				//fmt.Println(sync.OnlineElevators[1].ID)
				myElev.Orders[incomingMsg.Order.Floor][incomingMsg.Order.ButtonType].Status = UtilitiesTypes.Active
				sync.AddElevToMsgQueue(*myElev)
				switch myElev.State {
				case DOOR:
					fmt.Println("door")
					if !(myElev.Floor == incomingMsg.Order.Floor) {
						myElev.Orders[incomingMsg.Order.Floor][incomingMsg.Order.ButtonType].Status = UtilitiesTypes.Active
					} else {
						engineErrorTimeout.Reset(5 * time.Second)
						myElev.Orders[incomingMsg.Order.Floor][incomingMsg.Order.ButtonType].Status = UtilitiesTypes.Inactive
						elevio.SetDoorOpenLamp(true)
						doorTimeout.Reset(3 * time.Second)
					}
					break

				case MOVING:
					myElev.Orders[incomingMsg.Order.Floor][incomingMsg.Order.ButtonType].Status = UtilitiesTypes.Active
					break

				case IDLE:
					fmt.Println("før if")
					if myElev.Floor == incomingMsg.Order.Floor {
						fmt.Println("etter if")
						elevio.SetDoorOpenLamp(true)
						doorTimeout.Reset(3 * time.Second)
						myElev.Orders[incomingMsg.Order.Floor][incomingMsg.Order.ButtonType].Status = UtilitiesTypes.Inactive
						myElev.State = UtilitiesTypes.DOOR
					} else {
						engineErrorTimeout.Reset(3 * time.Second)
						myElev.Orders[incomingMsg.Order.Floor][incomingMsg.Order.ButtonType].Status = UtilitiesTypes.Active
						myElev.Dir = Requests.ChooseDirection(*myElev)
						elevio.SetMotorDirection(myElev.Dir)
						myElev.State = UtilitiesTypes.MOVING
					}
					break
				}
				Requests.SetAllCabLights(*myElev, numFloors, numButtons)
				sync.AddElevToMsgQueue(*myElev)
			}

		case <-doorTimeout.C:

			elevio.SetDoorOpenLamp(false)
			myElev.Dir = Requests.ChooseDirection(*myElev)
			if myElev.Dir == elevio.MD_Stop {
				myElev.State = IDLE
				engineErrorTimeout.Stop()
			} else {
				myElev.State = MOVING
				engineErrorTimeout.Reset(5 * time.Second)
				elevio.SetMotorDirection(myElev.Dir)
			}
		case <-engineErrorTimeout.C:
			fmt.Println("engine error")
			peerCh <- false
			sync.AddElevToMsgQueue(*myElev)

		}
	}

}
