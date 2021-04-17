package sync

import (
	"fmt"
	"time"

	"strconv"

	"../OrderDistributor"
	"../UtilitiesTypes"
	"../elevio"
	"../Network/network/peers"
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
	bestId := OrderDistributor.CostCalculator(OtherElevators, 1, 0)
	fmt.Println(bestId)

}

var MsgQueue = []UtilitiesTypes.Msg{}
var OnlineElevators = []UtilitiesTypes.Elevator{}

func CheckElevatorOnline(peerUpdateCh chan peers.PeerUpdate) {
	for {
		select {
		case p := <-peerUpdateCh:
			if len(OnlineElevators) != 0 {
				for i := 0; i < len(OnlineElevators); i++ {
					peer, _ := strconv.Atoi(p.Peers[i])
					if OnlineElevators[i].ID == peer {
						OnlineElevators[i].Online = true
						fmt.Println("online")
					} else {
						OnlineElevators[i].Online = false
						fmt.Println("offline")
					}

				}
			}
		}
	}

}

func ListContains(list []int, new int) bool {
	for i := 0; i < len(list); i++ {
		if list[i] == new {
			return true
		}
	}
	return false
}

func ContainsID(list []UtilitiesTypes.Elevator, new int) bool {
	for i := 0; i < len(list); i++ {
		if list[i].ID == new {
			return true
		}
	}
	return false
}

var iter = 0
var Message UtilitiesTypes.Msg

var (
	numPeers            = 2
	receivedMsg         []int
	LastIncomingMessage UtilitiesTypes.Msg
)

func AddHallOrderToMsgQueue(myElev UtilitiesTypes.Elevator, btnFloor int, btnType elevio.ButtonType) {
	iter++
	fmt.Println(OnlineElevators, "før cost")
	bestId := OrderDistributor.CostCalculator(OnlineElevators, btnFloor, btnType)
	for i := 0; i < len(OnlineElevators); i++ {
		OnlineElevators[i].Orders[btnFloor][btnType].Status = UtilitiesTypes.Inactive
	}
	fmt.Println(OnlineElevators, "etter cost")
	order := UtilitiesTypes.Order{Floor: btnFloor, ButtonType: int(btnType), Status: UtilitiesTypes.Active, Finished: false}
	Message.MsgID = iter
	Message.Elevator = myElev
	Message.IsNewOrder = true
	Message.Order = order
	Message.NewOrderTakerID = bestId
	Message.IsReceived = false
	Message.LocalID = myElev.ID
	MsgQueue = append(MsgQueue, Message)
	fmt.Println(Message.Elevator.Orders)
	fmt.Println("hall order")
	fmt.Println(bestId)

}

func UpdateHallLights() {

	for i := 0; i < len(OnlineElevators); i++ {
		for f := 0; f < UtilitiesTypes.NumFloors; f++ {
			if OnlineElevators[i].Orders[f][elevio.BT_HallUp].Status == UtilitiesTypes.Active {
				elevio.SetButtonLamp(elevio.BT_HallUp, f, true)
			}

			if OnlineElevators[i].Orders[f][elevio.BT_HallDown].Status == UtilitiesTypes.Active {
				elevio.SetButtonLamp(elevio.BT_HallDown, f, true)
			}

		}
	}
	for f := 0; f < UtilitiesTypes.NumFloors; f++ {
		for b := 0; b < 2; b++ {
			number := 0
			for i := 0; i < len(OnlineElevators); i++ {
				if OnlineElevators[i].Orders[f][b].Status == UtilitiesTypes.Inactive {
					number++
				}
			}
			if len(OnlineElevators) == number {
				elevio.SetButtonLamp(elevio.ButtonType(b), f, false)
			}

		}
	}
	time.Sleep(10 * time.Millisecond)

}

func AddElevToMsgQueue(myElev UtilitiesTypes.Elevator) {
	iter++
	Message.MsgID = iter
	Message.Elevator = myElev
	Message.IsNewOrder = false
	Message.IsReceived = false
	Message.LocalID = myElev.ID
	MsgQueue = append(MsgQueue, Message)
}

func SendMessage(msgChan UtilitiesTypes.MsgChan) {
	for {
		if !(len(MsgQueue) == 0) {
			msg := MsgQueue[0]
			msgChan.SendChan <- msg
			if len(receivedMsg) >= numPeers {
				MsgQueue = MsgQueue[1:]
				receivedMsg = receivedMsg[:0]
			} else {
				time.Sleep(10 * time.Millisecond)
			}

		}
		time.Sleep(4 * time.Millisecond) // Coopratrive routine
	}
}

// hvis timeren har gått ut, og vi må regne ut kostfunksjon på nytt hvis vi ikke får bekreftelse fra heisen som skulle ta ordren
// Alle hall orders til heisen som ikke lenger er i Peers må noen andre heiser ta ordrene.

func ConfirmationMessage(incomingMsg UtilitiesTypes.Msg, myElev UtilitiesTypes.Elevator, msgChan UtilitiesTypes.MsgChan) {
	var ConMessage UtilitiesTypes.Msg
	ConMessage.IsReceived = true
	ConMessage.IsNewOrder = false
	ConMessage.LocalID = myElev.ID
	ConMessage.MsgID = incomingMsg.MsgID
	msgChan.SendChan <- ConMessage
	time.Sleep(2 * time.Millisecond)
}

func Run(incomingMsg UtilitiesTypes.Msg, myElev UtilitiesTypes.Elevator, msgChan UtilitiesTypes.MsgChan) {
	UpdateHallLights()
	//fmt.Println(incomingMsg.Elevator.ID)
	//for i := 0; i < len(OnlineElevators); i++ {
	//fmt.Println(OnlineElevators[i].ID)
	//fmt.Println("---------", i)
	//}
	if !(incomingMsg.LocalID == myElev.ID) {
		if incomingMsg.IsReceived {
			if !ListContains(receivedMsg, incomingMsg.LocalID) {
				receivedMsg = append(receivedMsg, incomingMsg.LocalID)
				if len(receivedMsg) >= numPeers {
					// stoppe timer??

				}
			}
			// hvis timeren har gått ut, og vi må regne ut kostfunksjon på nytt hvis vi ikke får bekreftelse fra heisen som skulle ta ordren
			// Alle hall orders til heisen som ikke lenger er i Peers må noen andre heiser ta ordrene.
			//
		} else {
			ConfirmationMessage(incomingMsg, myElev, msgChan)

		}
	}
	if !(incomingMsg.IsReceived) && !(incomingMsg.IsNewOrder) {

		if !(LastIncomingMessage.MsgID == incomingMsg.MsgID && LastIncomingMessage.LocalID == incomingMsg.LocalID) {
			LastIncomingMessage.MsgID = incomingMsg.MsgID
			LastIncomingMessage.LocalID = incomingMsg.LocalID
			if len(OnlineElevators) != 0 {
				if ContainsID(OnlineElevators, incomingMsg.LocalID) {
					for i := 0; i < len(OnlineElevators); i++ {
						//fmt.Println(OnlineElevators[i].ID)
						//fmt.Println(OnlineElevators[i].Orders, "\n")
						if OnlineElevators[i].ID == incomingMsg.Elevator.ID {
							OnlineElevators[i] = incomingMsg.Elevator
							//fmt.Println("etter")
							//fmt.Println(OnlineElevators[i].ID)
							//fmt.Println(OnlineElevators[i].Orders, "\n")
						}
					}
				} else if !ContainsID(OnlineElevators, incomingMsg.LocalID) {
					if incomingMsg.LocalID != 0 {
						OnlineElevators = append(OnlineElevators, incomingMsg.Elevator)
						//fmt.Println(OnlineElevators)
						//fmt.Println(incomingMsg.LocalID)
					}
				}
			} else {
				OnlineElevators = append(OnlineElevators, myElev)
			}

		}
	}
	//fmt.Println(OnlineElevators)
	//fmt.Println("her kommer de på nytt")
}

func ShouldITake(incomingMsg UtilitiesTypes.Msg, myElev UtilitiesTypes.Elevator) bool {
	shouldITake := false
	if incomingMsg.IsNewOrder && !incomingMsg.IsReceived {
		if !(LastIncomingMessage.MsgID == incomingMsg.MsgID && LastIncomingMessage.LocalID == incomingMsg.LocalID) {
			LastIncomingMessage.MsgID = incomingMsg.MsgID
			LastIncomingMessage.LocalID = incomingMsg.LocalID
			if !ContainsID(OnlineElevators, incomingMsg.LocalID) {
				OnlineElevators = append(OnlineElevators, incomingMsg.Elevator)
			}

			for i := 0; i < len(OnlineElevators); i++ {
				fmt.Println(OnlineElevators[i].ID)
				fmt.Println(OnlineElevators[i].Orders, "\n")
				if OnlineElevators[i].ID == incomingMsg.Elevator.ID {
					OnlineElevators[i] = incomingMsg.Elevator
					fmt.Println("etter")
					fmt.Println(OnlineElevators[i].ID)
					fmt.Println(OnlineElevators[i].Orders, "\n")

				}

				if incomingMsg.NewOrderTakerID == myElev.ID {
					shouldITake = true

				}
			}

		}
	}
	//fmt.Println(OnlineElevators)
	//fmt.Println("her kommer de på nytt")
	return shouldITake
}
