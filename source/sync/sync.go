package sync

import (
	"fmt"

	"../OrderDistributor"
	"../UtilitiesTypes"
)

func Test() {

	Heis1 := UtilitiesTypes.Elevator{ID: 1, Dir: 0, Floor: 1, State: 3}
	Heis2 := UtilitiesTypes.Elevator{ID: 2, Dir: 0, Floor: 3, State: 3}
	Heis3 := UtilitiesTypes.Elevator{ID: 3, Dir: 0, Floor: 2, State: 3}
	Heis2.Orders[0][0].Status = UtilitiesTypes.Active
	Heis1.Orders[2][2].Status = UtilitiesTypes.Active
	Heis3.Orders[3][1].Status = UtilitiesTypes.Active
	Heis1.Orders[0][2].Status = UtilitiesTypes.Active

	var OtherElevators = []UtilitiesTypes.Elevator{Heis1, Heis2}
	best := OrderDistributor.CostCalculator(Heis3, OtherElevators, 1, 0)
	fmt.Println(best)

}
