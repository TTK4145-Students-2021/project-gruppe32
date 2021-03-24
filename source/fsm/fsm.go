package fsm

import (
	"fmt"

	"../Requests"
	"../UtilitiesTypes"
	"../elevio"
)

const numFloors = 4

const numButtons = 3

type State int

const (
	INIT   State = 0
	IDLE         = 1
	MOVING       = 2
	DOOR         = 3
)

var state State

func OnInitBetweenFloors(myElev UtilitiesTypes.Elevator) {
	elevio.SetMotorDirection(elevio.MD_Down)
	myElev.Dir = elevio.MD_Down
	myElev.State = UtilitiesTypes.MOVING
}

func OnRequestButtonPress(myElev *UtilitiesTypes.Elevator, btnFloor int, btnType elevio.ButtonType) {
	switch myElev.State {
	case DOOR:
		if myElev.Floor == btnFloor {
			Requests.SetStartTime()
		} else {
			myElev.Orders[btnFloor][btnType].Status = UtilitiesTypes.Active
			fmt.Println(myElev.State)
		}
		break

	case MOVING:
		myElev.Orders[btnFloor][btnType].Status = UtilitiesTypes.Active
		break

	case IDLE:
		if myElev.Floor == btnFloor {
			elevio.SetDoorOpenLamp(true)
			Requests.SetStartTime()
			myElev.State = UtilitiesTypes.DOOR
		} else {
			myElev.Orders[btnFloor][btnType].Status = UtilitiesTypes.Active
			myElev.Dir = Requests.ChooseDirection(*myElev)
			elevio.SetMotorDirection(myElev.Dir)
			myElev.State = UtilitiesTypes.MOVING
		}
		break
	}
	Requests.SetAllLights(*myElev, numFloors, numButtons)

}

func OnFloorArrival(myElev *UtilitiesTypes.Elevator, newFloor int) {
	myElev.Floor = newFloor

	elevio.SetFloorIndicator(myElev.Floor)

	switch myElev.State {
	case MOVING:
		if Requests.ShouldStop(*myElev) == true {
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevio.SetDoorOpenLamp(true)
			Requests.ClearAtCurrentFloor(myElev, numFloors, numButtons)
			Requests.SetStartTime()
			Requests.SetAllLights(*myElev, numFloors, numButtons)
			myElev.State = UtilitiesTypes.DOOR
		}
		break

	default:
		break
	}
}

func OnDoorTimeout(myElev *UtilitiesTypes.Elevator) {
	switch myElev.State {
	case DOOR:
		myElev.Dir = Requests.ChooseDirection(*myElev)
		elevio.SetDoorOpenLamp(false)
		elevio.SetMotorDirection(myElev.Dir)

		if myElev.Dir == elevio.MD_Stop {
			myElev.State = UtilitiesTypes.IDLE
		} else {
			myElev.State = UtilitiesTypes.MOVING
		}

		break
	default:
		break
	}
}

/*func FsmFunction(drv_buttons chan elevio.ButtonEvent, drv_floors chan int, drv_obstr chan bool, drv_stop chan bool){

	noOrder := UtilitiesTypes.Order{Floor: -1, ButtonType: -1, Status: -1, Finished: false, Confirmed: false}
	myElev := UtilitiesTypes.Elevator{Dir: 0, Floor: -1, State: 0, Orders: [numFloors][numButtons]noOrder}
	f:= 0
	currentOrder := elevio.ButtonEvent{Floor: -1, Button: -1}
	d := elevio.MD_Stop


	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)
	go Requests.UpdateLights(drv_buttons)

	state = INIT
	for {
		time.Sleep(20 * time.Millisecond)
		switch state{
		case INIT:
			select{
			case f = <- drv_floors:
				if f > 0 {
					d = elevio.MD_Stop
					myElevator.Floor = f
					state = IDLE
				}
				//else {
					//d = elevio.MD_Up
					//elevio.SetMotorDirection(d)
					//break
				//}
			}




		case IDLE:

			select{
			case currentOrder = <- drv_buttons:
				UtilitiesTypes.setOrder(myElevator, currentOrder.Floor,currentOrder.Button, 1, true, false)
				if Requests.RequestAbove(myElevator, numFloors, numButtons) {
					d = elevio.MD_Up
					elevio.SetMotorDirection(d)
					state = MOVING
				} else if Requests.RequestBelow(myElevator, numFloors, numButtons) {
					d = elevio.MD_Down
					elevio.SetMotorDirection(d)
					state = MOVING
				}



		}


		case MOVING:
			select{
			case currentOrder = <- drv_buttons:
				UtilitiesTypes.setOrder(myElevator, currentOrder.Floor,currentOrder.Button, 1, true, false)
				if Requests.ShouldStop(myElevator){
					d = elevio.MD_Stop
					elevio.SetMotorDirection(d)
					state = DOOR
				}
			}

		case DOOR:
	}
}
}


/*
for {
	select {
	case a := <- drv_buttons:
		fmt.Printf("%+v\n", a)
		elevio.SetButtonLamp(a.Button, a.Floor, true)

	case a := <- drv_floors:
		fmt.Printf("%+v\n", a)
		if a == numFloors-1 {
			d = elevio.MD_Down
		} else if a == 0 {
			d = elevio.MD_Up
		}
		elevio.SetMotorDirection(d)


	case a := <- drv_obstr:
		fmt.Printf("%+v\n", a)
		if a {
			elevio.SetMotorDirection(elevio.MD_Stop)
		} else {
			elevio.SetMotorDirection(d)
		}

	case a := <- drv_stop:
		fmt.Printf("%+v\n", a)
		for f := 0; f < numFloors; f++ {
			for b := elevio.ButtonType(0); b < 3; b++ {
				elevio.SetButtonLamp(b, f, false)
			}
		}
	}
}
*/
