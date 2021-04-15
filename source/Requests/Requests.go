package Requests

import (
	"time"

	"../UtilitiesTypes"
	"../elevio"
)

const numFloors = 4

const numButtons = 3

func UpdateLights(button chan elevio.ButtonEvent) {
	for {
		select {
		case a := <-button:
			elevio.SetButtonLamp(a.Button, a.Floor, true)
		}
	}
}

func SetAllCabLights(elev UtilitiesTypes.Elevator, numFloors int, numButtons int) {
	for floor := 0; floor < numFloors; floor++ {
		active := elev.Orders[floor][elevio.BT_Cab].Status == UtilitiesTypes.Active
		elevio.SetButtonLamp(elevio.ButtonType(elevio.BT_Cab), floor, active)
	}
}

func ClearAllLights(numFloors int, numButtons int) {
	for floor := 0; floor < numFloors; floor++ {
		for btn := 0; btn < numButtons; btn++ {
			elevio.SetButtonLamp(elevio.ButtonType(btn), floor, false)
		}
	}
	for floor := 0; floor < numFloors; floor++ {
		elevio.SetFloorIndicator(floor)
	}
	elevio.SetDoorOpenLamp(false)
}

func RequestAbove(elev UtilitiesTypes.Elevator, numFloors int, numButtons int) bool {
	for f := elev.Floor + 1; f < numFloors; f++ {
		for b := 0; b < numButtons; b++ {
			if elev.Orders[f][b].Status == UtilitiesTypes.Active {
				return true
			}
		}

	}
	return false
}

func RequestBelow(elev UtilitiesTypes.Elevator, numFloors int, numButtons int) bool {
	for f := 0; f < elev.Floor; f++ {
		for b := 0; b < numButtons; b++ {
			if elev.Orders[f][b].Status == UtilitiesTypes.Active {
				return true
			}
		}

	}
	return false
}

func ShouldStop(elev UtilitiesTypes.Elevator) bool {
	switch elev.Dir {
	case elevio.MD_Down:
		if elev.Orders[elev.Floor][elevio.BT_HallDown].Status == UtilitiesTypes.Active {
			return true
		} else if elev.Orders[elev.Floor][elevio.BT_Cab].Status == UtilitiesTypes.Active {
			return true
		} else if !RequestBelow(elev, numFloors, numButtons) {
			return true
		}
	case elevio.MD_Up:
		if elev.Orders[elev.Floor][elevio.BT_HallUp].Status == UtilitiesTypes.Active {
			return true
		} else if elev.Orders[elev.Floor][elevio.BT_Cab].Status == UtilitiesTypes.Active {
			return true
		} else if !RequestAbove(elev, numFloors, numButtons) {
			return true
		}
	default:
		return true

	}
	return false
}

func ChooseDirection(elev UtilitiesTypes.Elevator) elevio.MotorDirection {
	switch elev.Dir {
	case elevio.MD_Up:
		if RequestAbove(elev, numFloors, numButtons) {
			return elevio.MD_Up
		} else if RequestBelow(elev, numFloors, numButtons) {
			return elevio.MD_Down
		}
		return elevio.MD_Stop
	case elevio.MD_Down:
		if RequestBelow(elev, numFloors, numButtons) {
			return elevio.MD_Down
		} else if RequestAbove(elev, numFloors, numButtons) {
			return elevio.MD_Up
		}
		return elevio.MD_Stop
	case elevio.MD_Stop:
		if RequestBelow(elev, numFloors, numButtons) {
			return elevio.MD_Down
		} else if RequestAbove(elev, numFloors, numButtons) {
			return elevio.MD_Up
		}
		return elevio.MD_Stop
	default:
		return elevio.MD_Stop
	}
}

func ClearAtCurrentFloor(elev *UtilitiesTypes.Elevator, numFloors int, numButtons int) {
	for btn := 0; btn < numButtons; btn++ {
		elev.Orders[elev.Floor][btn].Status = UtilitiesTypes.Inactive
		elev.Orders[elev.Floor][btn].Finished = true
	}
}

var startTime time.Time

func SetStartTime() {
	startTime = time.Now()
}

func GetStartTime() time.Time {
	return startTime
}

func TimeOut(seconds time.Duration, myElev UtilitiesTypes.Elevator) bool {
	if myElev.State == UtilitiesTypes.DOOR {
		seconds = seconds * time.Second
		begin := GetStartTime()
		difference := time.Now().Sub(begin)

		if difference >= seconds {
			return true
		}
	}
	return false
}
