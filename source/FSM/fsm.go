package FSM

import (
	"fmt"
	"time"

	eio "../ElevIO"
	Req "../Requests"
	"../Sync"
	UT "../UtilitiesTypes"
)

const (
	NumFloors  = UT.NumFloors
	NumButtons = UT.NumButtons
)

type State int

const (
	IDLE   = 1
	MOVING = 2
	DOOR   = 3
)

func FSM(msgChan UT.MsgChan, drv_buttons chan eio.ButtonEvent, drv_floors chan int, myElev *UT.Elevator, peerCh chan bool, drv_obstr chan bool) {
	for eio.GetFloor() == -1 {
		eio.SetMotorDirection(eio.MD_Down)
		myElev.Dir = eio.MD_Down
		myElev.State = UT.MOVING
		time.Sleep(4 * time.Millisecond)
	}

	doorTimeout := time.NewTimer(3 * time.Second)
	engineErrorTimeout := time.NewTimer(5 * time.Second)
	doorTimeout.Stop()
	engineErrorTimeout.Stop()
	for {
		Sync.UpdateHallLights()
		select {
		case btn := <-drv_buttons:
			if btn.Button == eio.BT_Cab {
				switch myElev.State {
				case DOOR:
					if !(myElev.Floor == btn.Floor) {
						myElev.Orders[btn.Floor][btn.Button].Status = UT.Active
					} else {
						engineErrorTimeout.Reset(5 * time.Second)
						myElev.Orders[btn.Floor][btn.Button].Status = UT.Inactive
						eio.SetDoorOpenLamp(true)
						doorTimeout.Reset(3 * time.Second)
					}

				case MOVING:
					myElev.Orders[btn.Floor][btn.Button].Status = UT.Active

				case IDLE:
					if myElev.Floor == btn.Floor {
						eio.SetDoorOpenLamp(true)
						doorTimeout.Reset(3 * time.Second)
						myElev.Orders[btn.Floor][btn.Button].Status = UT.Inactive
						myElev.State = UT.DOOR
					} else {
						engineErrorTimeout.Reset(3 * time.Second)
						myElev.Orders[btn.Floor][btn.Button].Status = UT.Active
						myElev.Dir = Req.ChooseDirection(*myElev)
						eio.SetMotorDirection(myElev.Dir)
						myElev.State = UT.MOVING
					}
				}
				Req.SetAllCabLights(*myElev, NumFloors, NumButtons)
				Sync.AddElevToMsgQueue(*myElev)
			} else {
				Sync.AddHallOrderToMsgQueue(*myElev, btn.Floor, btn.Button)
			}
		case newFloor := <-drv_floors:
			myElev.Floor = newFloor
			myElev.MotorStop = false
			peerCh <- true
			engineErrorTimeout.Reset(5 * time.Second)

			eio.SetFloorIndicator(myElev.Floor)

			if Req.ShouldStop(*myElev) {
				eio.SetMotorDirection(eio.MD_Stop)
				myElev.Dir = eio.MD_Stop
				eio.SetDoorOpenLamp(true)
				Req.ClearAtCurrentFloor(myElev, NumFloors, NumButtons)
				doorTimeout.Reset(3 * time.Second)
				engineErrorTimeout.Stop()
				Req.SetAllCabLights(*myElev, NumFloors, NumButtons)
				myElev.State = UT.DOOR
			} else if myElev.State == MOVING {
				engineErrorTimeout.Reset(3 * time.Second)
			}

			Sync.AddElevToMsgQueue(*myElev)
			Req.SetAllCabLights(*myElev, NumFloors, NumButtons)

		case obstruction := <-drv_obstr:
			if obstruction && myElev.State == DOOR {
				doorTimeout.Stop()
				engineErrorTimeout.Reset(3 * time.Second)
			} else if !obstruction && myElev.State == DOOR {
				engineErrorTimeout.Stop()
				doorTimeout.Reset(3 * time.Second)

			}

		case incomingMsg := <-msgChan.RecChan:
			Sync.HandleIncomingMessages(incomingMsg, myElev, msgChan)
			if Sync.HandleNewOrder(incomingMsg, *myElev) {
				Req.SetAllCabLights(*myElev, NumFloors, NumButtons)
				myElev.Orders[incomingMsg.Order.Floor][incomingMsg.Order.ButtonType].Status = UT.Active
				Sync.AddElevToMsgQueue(*myElev)

				switch myElev.State {

				case DOOR:
					if !(myElev.Floor == incomingMsg.Order.Floor) {
						myElev.Orders[incomingMsg.Order.Floor][incomingMsg.Order.ButtonType].Status = UT.Active
					} else {
						engineErrorTimeout.Reset(5 * time.Second)
						myElev.Orders[incomingMsg.Order.Floor][incomingMsg.Order.ButtonType].Status = UT.Inactive
						eio.SetDoorOpenLamp(true)
						doorTimeout.Reset(3 * time.Second)
					}

				case MOVING:
					myElev.Orders[incomingMsg.Order.Floor][incomingMsg.Order.ButtonType].Status = UT.Active

				case IDLE:
					if myElev.Floor == incomingMsg.Order.Floor {
						eio.SetDoorOpenLamp(true)
						doorTimeout.Reset(3 * time.Second)
						myElev.Orders[incomingMsg.Order.Floor][incomingMsg.Order.ButtonType].Status = UT.Inactive
						myElev.State = UT.DOOR
					} else {
						engineErrorTimeout.Reset(3 * time.Second)
						myElev.Orders[incomingMsg.Order.Floor][incomingMsg.Order.ButtonType].Status = UT.Active
						myElev.Dir = Req.ChooseDirection(*myElev)
						eio.SetMotorDirection(myElev.Dir)
						myElev.State = UT.MOVING
					}
				}
				Req.SetAllCabLights(*myElev, NumFloors, NumButtons)
				Sync.AddElevToMsgQueue(*myElev)
			}

		case <-doorTimeout.C:

			eio.SetDoorOpenLamp(false)
			myElev.Dir = Req.ChooseDirection(*myElev)
			if myElev.Dir == eio.MD_Stop {
				myElev.State = IDLE
				engineErrorTimeout.Stop()
			} else {
				myElev.State = MOVING
				engineErrorTimeout.Reset(5 * time.Second)
				eio.SetMotorDirection(myElev.Dir)
			}
		case <-engineErrorTimeout.C:
			fmt.Println("engine error")
			peerCh <- false
			Sync.AddElevToMsgQueue(*myElev)
			time.Sleep(1 * time.Second)
			for f := 0; f < NumFloors; f++ {
				for btn := 0; btn < NumButtons-1; btn++ {
					myElev.Orders[f][btn].Status = UT.Inactive
				}
			}
			Sync.AddElevToMsgQueue(*myElev)
		}

	}
}
