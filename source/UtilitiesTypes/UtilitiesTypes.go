package UtilitiesTypes

import "../elevio"

const NumFloors = 4

const NumButtons = 3

var myElevator Elevator

type State int

const (
	INIT   State = 0
	IDLE         = 1
	MOVING       = 2
	DOOR         = 3
)

type Order struct {
	Floor      int
	ButtonType int
	Status     OrderStatus
	Finished   bool
}

type OrderStatus int

const (
	OrderTimeout OrderStatus = -2
	Inactive                 = 0
	Pending                  = 2
	Active                   = 1
)

type Elevator struct {
	ID     int
	Dir    elevio.MotorDirection
	Floor  int
	State  State
	Orders [NumFloors][NumButtons]Order
}

type Msg struct {
	Elevator        Elevator
	IsNewOrder      bool
	Order           Order
	NewOrderTakerID int
	IsReceived      bool
	MsgID           int
	LocalID         int
}

type MsgChan struct {
	SendChan chan Msg
	RecChan  chan Msg
}

func getOrder(floor int, buttonType int) Order {
	return myElevator.Orders[floor][buttonType]
}

func getOrderList() [NumFloors][NumButtons]Order {
	return myElevator.Orders
}

func SetOrder(myElevator *Elevator, floor int, buttonType elevio.ButtonType, status OrderStatus, finished bool) {
	myElevator.Orders[floor][buttonType].Status = status
	myElevator.Orders[floor][buttonType].Finished = finished

}

func setElevator(floor int, dir elevio.MotorDirection, state State, order [NumFloors][NumButtons]Order) Elevator {
	myElevator.Floor = floor
	myElevator.Dir = dir
	myElevator.State = state
	myElevator.Orders = order
	return myElevator
}
