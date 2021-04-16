package OrderDistributor

import (
	"../Requests"
	"../UtilitiesTypes"
	"../elevio"
)

var TRAVEL_TIME = 2500
var DOOR_OPEN_TIME = 3000

const numFloors = 4
const numButtons = 3

func TimeToIdle(myElev UtilitiesTypes.Elevator) int {
	duration := 0

	switch myElev.State {
	case UtilitiesTypes.IDLE:
		myElev.Dir = Requests.ChooseDirection(myElev)
		if myElev.Dir == elevio.MD_Stop {
			return duration
		}
		break

	case UtilitiesTypes.MOVING:
		duration += TRAVEL_TIME / 2
		myElev.Floor += int(myElev.Dir)
		break

	case UtilitiesTypes.DOOR:
		duration -= DOOR_OPEN_TIME / 2
	}

	for {
		if Requests.ShouldStop(myElev) {
			Requests.ClearAtCurrentFloor(&myElev, numFloors, numButtons)
			duration += DOOR_OPEN_TIME
			myElev.Dir = Requests.ChooseDirection(myElev)
			if myElev.Dir == elevio.MD_Stop {
				return duration
			}
		}
		myElev.Floor += int(myElev.Dir)
		duration += TRAVEL_TIME
	}
}

func CostCalculator(allElevators [UtilitiesTypes.NumElevs]UtilitiesTypes.Elevator, btnFloor int, btnType elevio.ButtonType) int {

	UtilitiesTypes.SetOrder(&allElevators[0], btnFloor, btnType, UtilitiesTypes.Active, false)
	cost := TimeToIdle(allElevators[0])

	bestElevator := allElevators[0]

	for i := 1; i < len(allElevators); i++ {
		UtilitiesTypes.SetOrder(&allElevators[i], btnFloor, btnType, UtilitiesTypes.Active, false)

		if TimeToIdle(allElevators[i]) < cost {
			cost = TimeToIdle(allElevators[i])
			bestElevator = allElevators[i]
		}
	}

	return bestElevator.ID
}
