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
	INIT   State = 0
	IDLE         = 1
	MOVING       = 2
	DOOR         = 3
)




func FsmElevator(ch UtilitiesTypes.FsmChan, id int){
	elev := UtilitiesTypes.Elevator{
		State: IDLE,
		Dir: elevio.MD_Stop,
		Floor: elevio.GetFloor(),
		ID: id,

	}
	doorTimeout := time.NewTimer(3*time.Second)
	engineErrorTimeout := time.NewTimer(3*time.Second)
	doorTimeout.Stop()
	engineErrorTimeout.Stop()
	fmt.Println("inni fsm")


	for{
		select{
		case newButtonPress := <- ch.NewButtonPress:
			if newButtonPress.Button == elevio.BT_Cab {

				switch elev.State{
				case DOOR:
					//fmt.Println("door")
					if !(elev.Floor == newButtonPress.Floor) {
						elev.Orders[newButtonPress.Floor][newButtonPress.Button].Status = UtilitiesTypes.Active
					} else {
						elevio.SetDoorOpenLamp(true)
						doorTimeout.Reset(3*time.Second)
					}
					break
			
				case MOVING:
					elev.Orders[newButtonPress.Floor][newButtonPress.Button].Status = UtilitiesTypes.Active
					break
			
				case IDLE:
					if elev.Floor == newButtonPress.Floor {
						elevio.SetDoorOpenLamp(true)
						doorTimeout.Reset(3*time.Second)
						elev.Orders[newButtonPress.Floor][newButtonPress.Button].Status = UtilitiesTypes.Inactive
						elev.State = DOOR
					} else {
						elev.Orders[newButtonPress.Floor][newButtonPress.Button].Status = UtilitiesTypes.Active
						elev.Dir = Requests.ChooseDirection(elev)
						elevio.SetMotorDirection(elev.Dir)
						elev.State = MOVING
					}
					break
				}

				}else{
					fmt.Println("legger til hallorder")
					sync.AddHallOrderToMsgQueue(elev,newButtonPress.Floor,newButtonPress.Button)
				}
				Requests.SetAllCabLights(elev, numFloors, numButtons)
				ch.Elev <- elev
		
		case incomingOrder := <- ch.NewOrder:

			switch elev.State{
			case DOOR:
				//fmt.Println("door")
				if !(elev.Floor == incomingOrder.Floor) {
					elev.Orders[incomingOrder.Floor][incomingOrder.Button].Status = UtilitiesTypes.Active
				} else {
					elevio.SetDoorOpenLamp(true)
					doorTimeout.Reset(3*time.Second)
				}
				break
		
			case MOVING:
				elev.Orders[incomingOrder.Floor][incomingOrder.Button].Status = UtilitiesTypes.Active
				break
		
			case IDLE:
				if elev.Floor == incomingOrder.Floor {
					elevio.SetDoorOpenLamp(true)
					doorTimeout.Reset(3*time.Second)
					elev.Orders[incomingOrder.Floor][incomingOrder.Button].Status = UtilitiesTypes.Inactive
					elev.State = DOOR
				} else {
					elev.Orders[incomingOrder.Floor][incomingOrder.Button].Status = UtilitiesTypes.Active
					elev.Dir = Requests.ChooseDirection(elev)
					elevio.SetMotorDirection(elev.Dir)
					elev.State = MOVING
				}
				break
			}

			ch.Elev <- elev
		case elevfloor := <- ch.ArrivedAtFloor:
			elev.Floor = elevfloor

			elevio.SetFloorIndicator(elev.Floor)

			switch elev.State {
			case MOVING:
				if Requests.ShouldStop(elev) {
					elevio.SetMotorDirection(elevio.MD_Stop)
					elevio.SetDoorOpenLamp(true)
					Requests.ClearAtCurrentFloor(&elev, numFloors, numButtons)
					doorTimeout.Reset(3*time.Second)
					Requests.SetAllCabLights(elev, numFloors, numButtons)
					elev.State = DOOR
			}
			break

			default:
			//assert("THIS SHOULD NOT BE CALLEd")
			break
			}
			ch.Elev <- elev
	case <-doorTimeout.C:
			
		elevio.SetDoorOpenLamp(false)
		elev.Dir = Requests.ChooseDirection(elev)
		if elev.Dir == elevio.MD_Stop{
			elev.State = IDLE
			engineErrorTimeout.Stop()
		}else{
			elev.State = MOVING
			engineErrorTimeout.Reset(3*time.Second)
			elevio.SetMotorDirection(elev.Dir)
		}
		ch.Elev <- elev
			}
		}

	
}

func DoorState(ch UtilitiesTypes.FsmChan) {
	elev := <- ch.Elev
	ch.Elev <- elev
	for {
		if Requests.TimeOut(3, elev) {
			fmt.Println("TimerOut")
			OnDoorTimeout(elev)
		}
	}
}
func OnDoorTimeout(myElev UtilitiesTypes.Elevator) {
	switch myElev.State {
	case DOOR:
		myElev.Dir = Requests.ChooseDirection(myElev)
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

			/*
			switch elev.State{

			case IDLE:
				elev.Dir = Requests.ChooseDirection(elev)
				elevio.SetMotorDirection(elev.Dir)
				if elev.Dir == elevio.MD_Stop{
					elev.State = DOOR
					elevio.SetDoorOpenLamp(true)
					doorTimeout.Reset(3*time.Second)
					
					//endre til orderFinisihed
					//Fjerne order fra kø

				} else{
					elev.State = MOVING
					//sjekke for motorstopp
				}
			case MOVING:
			case DOOR:
				if elev.Floor == newButtonPress.Floor{
					doorTimeout.Reset(3*time.Second)
					//endre til orderFinisihed
					//Fjerne order fra kø
				}
				}
				} else {
					sync.AddHallOrderToMsgQueue(elev, newButtonPress.Floor, newButtonPress.Button)
				}
			//case Undefined:
			//default:
			//	fmt.Println("Error")
			//}
			ch.Elev <- elev
		
		case newOrder := <- ch.NewOrder:
			switch elev.State{

			case IDLE:
				elev.Dir = Requests.ChooseDirection(elev)
				elevio.SetMotorDirection(elev.Dir)
				if elev.Dir == elevio.MD_Stop{
					elev.State = DOOR
					elevio.SetDoorOpenLamp(true)
					doorTimeout.Reset(3*time.Second)
					
					//endre til orderFinisihed
					//Fjerne order fra kø

				} else{
					elev.State = MOVING
					//sjekke for motorstopp
				}
			case MOVING:
			case DOOR:
				if elev.Floor == newOrder.Floor{
					doorTimeout.Reset(3*time.Second)
					//endre til orderFinisihed
					//Fjerne order fra kø
				}
			//case Undefined:
			//default:
			//	fmt.Println("Error")
			//}
			ch.Elev <- elev
			}




		case elevfloor := <- ch.ArrivedAtFloor:
			
			elev.Floor = elevfloor
			if Requests.ShouldStop(elev){
				//Finished = true
				elevio.SetDoorOpenLamp(true)
				//sjekk motorstopp
				elev.State = DOOR
				elevio.SetMotorDirection(elevio.MD_Stop)
				// sett på dørtimer i 3 sek
				//endre til orderFinisihed
					//Fjerne order fra kø
			}else if elev.State == MOVING{
				//sjekk motorstopp
			}
			ch.Elev <- elev

		case <-doorTimeout.C:
			
				elevio.SetDoorOpenLamp(false)
				elev.Dir = Requests.ChooseDirection(elev)
				if elev.Dir == elevio.MD_Stop{
					elev.State = IDLE
					engineErrorTimeout.Stop()
				}else{
					elev.State = MOVING
					engineErrorTimeout.Reset(3*time.Second)
					elevio.SetMotorDirection(elev.Dir)
				}
				ch.Elev <- elev

			case <-engineErrorTimeout.C:
				//elevio.SetMotorDir(STOP)
				//Elevator.State == Undefined
				//print at heisen har motorstopp
				//elevio.SetMotorDir(elev.Dir)
				ch.Elev <- elev
				engineErrorTimeout.Reset(5*time.Second)
				*/
		




//var state State
/*
func OnInitBetweenFloors(myElev *UtilitiesTypes.Elevator) {
	elevio.SetMotorDirection(elevio.MD_Down)
	myElev.Dir = elevio.MD_Down
	myElev.State = UtilitiesTypes.MOVING
}

func DoorState(myElev *UtilitiesTypes.Elevator) {
	for {
		if Requests.TimeOut(3, *myElev) {
			fmt.Println("TimerOut")
			OnDoorTimeout(myElev)
		}
	}
}

func OnRequestButtonPress(myElev *UtilitiesTypes.Elevator, btnFloor int, btnType elevio.ButtonType) {
	switch myElev.State {
	case DOOR:
		fmt.Println("door")
		if !(myElev.Floor == btnFloor) {
			myElev.Orders[btnFloor][btnType].Status = UtilitiesTypes.Active
		} else {
			elevio.SetDoorOpenLamp(true)
			Requests.SetStartTime()
		}
		break

	case MOVING:
		myElev.Orders[btnFloor][btnType].Status = UtilitiesTypes.Active
		break

	case IDLE:
		fmt.Println("før if")
		if myElev.Floor == btnFloor {
			fmt.Println("etter if")
			elevio.SetDoorOpenLamp(true)
			Requests.SetStartTime()
			myElev.Orders[btnFloor][btnType].Status = UtilitiesTypes.Inactive
			myElev.State = UtilitiesTypes.DOOR
		} else {
			myElev.Orders[btnFloor][btnType].Status = UtilitiesTypes.Active
			myElev.Dir = Requests.ChooseDirection(*myElev)
			elevio.SetMotorDirection(myElev.Dir)
			myElev.State = UtilitiesTypes.MOVING
		}
		break
	}
	Requests.SetAllCabLights(*myElev, numFloors, numButtons)

}

func OnFloorArrival(myElev *UtilitiesTypes.Elevator, newFloor int) {
	myElev.Floor = newFloor

	elevio.SetFloorIndicator(myElev.Floor)

	switch myElev.State {
	case MOVING:
		if Requests.ShouldStop(*myElev) {
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevio.SetDoorOpenLamp(true)
			Requests.ClearAtCurrentFloor(myElev, numFloors, numButtons)
			Requests.SetStartTime()
			Requests.SetAllCabLights(*myElev, numFloors, numButtons)
			myElev.State = UtilitiesTypes.DOOR
		}
		break

	default:
		//assert("THIS SHOULD NOT BE CALLEd")
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
*/







//GAMMEL UTDELT

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
}var state State
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