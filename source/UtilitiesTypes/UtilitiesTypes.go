package UtilitiesTypes

import eio "../ElevIO"

const NumFloors = 4
const NumButtons = 3

var myElevator Elevator

type State int

const (
	IDLE   State = 1
	MOVING       = 2
	DOOR         = 3
)

type Order struct {
	Floor      int
	ButtonType int
	Status     OrderStatus
}

type OrderStatus int

const (
	Inactive OrderStatus = 0
	Active               = 1
)

type Elevator struct {
	ID        int
	Dir       eio.MotorDirection
	Floor     int
	State     State
	Orders    [NumFloors][NumButtons]Order
	MotorStop bool
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
