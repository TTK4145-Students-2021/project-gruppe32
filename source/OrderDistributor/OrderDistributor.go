package OrderDistributor

import (
	Req "../Requests"
	UT "../UtilitiesTypes"
	eio "../ElevIO"
)

var TRAVEL_TIME = 2500
var DOOR_OPEN_TIME = 3000

const (
	NumButtons = UT.NumButtons
	NumFloors  = UT.NumFloors
)

func TimeToIdle(myElev UT.Elevator) int {
	duration := 0

	switch myElev.State {
	case UT.IDLE:
		myElev.Dir = Req.ChooseDirection(myElev)
		if myElev.Dir == eio.MD_Stop {
			return duration
		}
		break

	case UT.MOVING:
		duration += TRAVEL_TIME / 2
		if !(myElev.Floor == 0 && myElev.Dir == -1) {
			myElev.Floor += int(myElev.Dir)
		}
		break

	case UT.DOOR:
		duration -= DOOR_OPEN_TIME / 2
	}

	for {
		if Req.ShouldStop(myElev) {
			Req.ClearAtCurrentFloor(&myElev, NumFloors, NumButtons)
			duration += DOOR_OPEN_TIME
			myElev.Dir = Req.ChooseDirection(myElev)
			if myElev.Dir == eio.MD_Stop {
				return duration
			}
		}
		if !(myElev.Floor == 0 && myElev.Dir == -1) {
			myElev.Floor += int(myElev.Dir)
		}
		duration += TRAVEL_TIME
	}
}

func CostCalculator(onlineElevators []UT.Elevator, btnFloor int, btnType eio.ButtonType, myElev UT.Elevator) int {
	if eio.GetFloor() == -1 {
		for j := 0; j < len(onlineElevators); j++ {
			if myElev.ID == onlineElevators[j].ID {
				onlineElevators = append(onlineElevators[:j], onlineElevators[j+1:]...)
			}
		}
	}
	onlineElevators[0].Orders[btnFloor][btnType].Status = UT.Active
	cost := TimeToIdle(onlineElevators[0])

	bestElevator := onlineElevators[0]

	if len(onlineElevators) > 0 {

		for i := 0; i < len(onlineElevators); i++ {
			onlineElevators[i].Orders[btnFloor][btnType].Status = UT.Active

			if TimeToIdle(onlineElevators[i]) < cost {
				cost = TimeToIdle(onlineElevators[i])
				bestElevator = onlineElevators[i]
			}
		}

	}

	bestID := bestElevator.ID

	return bestID

}
