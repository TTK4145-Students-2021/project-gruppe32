package main

import (
	//"fmt"

	"./Requests"
	"./UtilitiesTypes"
	"./elevio"
	"./fsm"
	"./sync"
)

var myElevator UtilitiesTypes.Elevator

func DoorState() {
	for {
		//fmt.Printf("Door state")
		if Requests.TimeOut(3) {
			//fmt.Printf("Timeout")
			fsm.OnDoorTimeout(&myElevator)
		}
	}
}

func main() {

	numFloors := 4
	numButtons := 3
	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	elevio.Init("localhost:15657", numFloors)
	myElevator.State = fsm.IDLE
	Requests.ClearAllLights(numFloors, numButtons)

	sync.Test()

	/*go DoorState()
	for {
		select {
		case a := <-drv_buttons:
			fsm.OnRequestButtonPress(&myElevator, a.Floor, a.Button)
			//fmt.Println(myElevator.State)
			//fmt.Println(myElevator.Orders[2][2].Status)
		case a := <-drv_floors:
			fsm.OnFloorArrival(&myElevator, a)
		}
	}





	//FsmFunction(drv_buttons, drv_floors, drv_obstr, drv_stop)
	//var noOrder UtilitiesTypes.Order
	//noOrder = UtilitiesTypes.Order{Floor: -1, ButtonType: -1, Status: -1, Finished: false}
	//myElev := UtilitiesTypes.Elevator{Dir: 0, Floor: -1, State: 0, Orders: [numFloors][numButtons]noOrder}
	//f:= 0
	//currentOrder := elevio.ButtonEvent{Floor: -1, Button: -1}

	/*
			for {
				select {
				case a := <-drv_buttons:
					fsm.OnRequestButtonPress(a.Floor, a.Button)

				case a := <-drv_floors:
					fsm.OnFloorArrival(a)

				case a := <-drv_stop:
					fmt.Printf("%+v\n", a)
					for f := 0; f < numFloors; f++ {
						for b := elevio.ButtonType(0); b < 3; b++ {
							elevio.SetButtonLamp(b, f, false)
						}
					}
				}
			}

		}

		/*var d elevio.MotorDirection = elevio.MD_Up
			//elevio.SetMotorDirection(d)

			drv_buttons := make(chan elevio.ButtonEvent)
			drv_floors := make(chan int)
			drv_obstr := make(chan bool)
			drv_stop := make(chan bool)

			go elevio.PollButtons(drv_buttons)
			go elevio.PollFloorSensor(drv_floors)
			go elevio.PollObstructionSwitch(drv_obstr)
			go elevio.PollStopButton(drv_stop)
			go orderhandler.UpdateLights(drv_buttons)

			for {
				select {
				//case a := <-drv_buttons:
					//orderhandler.UpdateLights(a)
					//fmt.Printf("%+v\n", a)
					//elevio.SetButtonLamp(a.Button, a.Floor, true)

				case a := <-drv_floors:
					fmt.Printf("%+v\n", a)
					if a == numFloors-1 {
						d = elevio.MD_Down
					} else if a == 0 {
						d = elevio.MD_Up
					}
					elevio.SetMotorDirection(d)

				case a := <-drv_obstr:
					fmt.Printf("%+v\n", a)
					if a {
						elevio.SetMotorDirection(elevio.MD_Stop)
					} else {
						elevio.SetMotorDirection(d)
					}

				case a := <-drv_stop:
					fmt.Printf("%+v\n", a)
					for f := 0; f < numFloors; f++ {
						for b := elevio.ButtonType(0); b < 3; b++ {
							elevio.SetButtonLamp(b, f, false)
						}
					}
				}
			}*/
}
