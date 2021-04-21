package Requests

import (
	UT "../UtilitiesTypes"
	eio "../ElevIO"
)

const (
	NumFloors  = UT.NumFloors
	NumButtons = UT.NumButtons
)

func SetAllCabLights(elev UT.Elevator, NumFloors int, NumButtons int) {
	for floor := 0; floor < NumFloors; floor++ {
		active := elev.Orders[floor][eio.BT_Cab].Status == UT.Active
		eio.SetButtonLamp(eio.ButtonType(eio.BT_Cab), floor, active)
	}
}

func ClearAllLights(NumFloors int, NumButtons int) {
	for floor := 0; floor < NumFloors; floor++ {
		for btn := 0; btn < NumButtons; btn++ {
			eio.SetButtonLamp(eio.ButtonType(btn), floor, false)
		}
	}
	for floor := 0; floor < NumFloors; floor++ {
		eio.SetFloorIndicator(floor)
	}
	eio.SetDoorOpenLamp(false)
}

func requestAbove(elev UT.Elevator, NumFloors int, NumButtons int) bool {
	for f := elev.Floor + 1; f < NumFloors; f++ {
		for b := 0; b < NumButtons; b++ {
			if elev.Orders[f][b].Status == UT.Active {
				return true
			}
		}

	}
	return false
}

func requestBelow(elev UT.Elevator, NumFloors int, NumButtons int) bool {
	for f := 0; f < elev.Floor; f++ {
		for b := 0; b < NumButtons; b++ {
			if elev.Orders[f][b].Status == UT.Active {
				return true
			}
		}

	}
	return false
}

func ShouldStop(elev UT.Elevator) bool {
	switch elev.Dir {
	case eio.MD_Down:
		return elev.Orders[elev.Floor][eio.BT_HallDown].Status == UT.Active || elev.Orders[elev.Floor][eio.BT_Cab].Status == UT.Active || !requestBelow(elev, NumFloors, NumButtons)
	case eio.MD_Up:
		return elev.Orders[elev.Floor][eio.BT_HallUp].Status == UT.Active || elev.Orders[elev.Floor][eio.BT_Cab].Status == UT.Active || !requestAbove(elev, NumFloors, NumButtons)
	case eio.MD_Stop:
	default:
		return true

	}
	return true
}

func ChooseDirection(elev UT.Elevator) eio.MotorDirection {
	switch elev.Dir {
	case eio.MD_Up:
		if requestAbove(elev, NumFloors, NumButtons) {
			return eio.MD_Up
		} else if requestBelow(elev, NumFloors, NumButtons) {
			return eio.MD_Down
		}
		return eio.MD_Stop
	case eio.MD_Down:
		if requestBelow(elev, NumFloors, NumButtons) {
			return eio.MD_Down
		} else if requestAbove(elev, NumFloors, NumButtons) {
			return eio.MD_Up
		}
		return eio.MD_Stop
	case eio.MD_Stop:
		if requestBelow(elev, NumFloors, NumButtons) {
			return eio.MD_Down
		} else if requestAbove(elev, NumFloors, NumButtons) {
			return eio.MD_Up
		}
		return eio.MD_Stop
	default:
		return eio.MD_Stop
	}
}

func ClearAtCurrentFloor(elev *UT.Elevator, NumFloors int, NumButtons int) {
	for btn := 0; btn < NumButtons; btn++ {
		elev.Orders[elev.Floor][btn].Status = UT.Inactive
	}
}
