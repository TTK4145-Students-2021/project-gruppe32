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

func CostCalculator(onlineElevators []UtilitiesTypes.Elevator, btnFloor int, btnType elevio.ButtonType) int {

	onlineElevators[0].Orders[btnFloor][btnType].Status = UtilitiesTypes.Active
	cost := TimeToIdle(onlineElevators[0])

	bestElevator := onlineElevators[0]

	if len(onlineElevators) > 0 {

		for i := 0; i < len(onlineElevators); i++ {
			onlineElevators[i].Orders[btnFloor][btnType].Status = UtilitiesTypes.Active

			if TimeToIdle(onlineElevators[i]) < cost {
				cost = TimeToIdle(onlineElevators[i])
				bestElevator = onlineElevators[i]
			}
		}

	}
	bestID := bestElevator.ID

	return bestID
}
