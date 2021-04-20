package sync

import (
	"fmt"
	"time"

	"strconv"

	"../Network/network/peers"
	OD "../OrderDistributor"
	UT "../UtilitiesTypes"
	eio "../elevio"
)

var MsgQueue = []UT.Msg{}
var AllElevators = []UT.Elevator{}
var OnlineIds = []int{}
var iter = 0
var Message UT.Msg

var (
	receivedMsg         []int
	LastIncomingMessage UT.Msg
)

const (
	NumFloors  = UT.NumFloors
	NumButtons = UT.NumButtons
)

func UpdateOnlineIds(peerUpdateCh chan peers.PeerUpdate, myElev UT.Elevator) {
	for {
		select {
		case p := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)
			for i := 0; i < len(p.Peers); i++ {
				peer, _ := strconv.Atoi(p.Peers[i])
				if len(OnlineIds) != 0 {
					if !ListContains(OnlineIds, peer) {
						OnlineIds = append(OnlineIds, peer)

					}
				}
			}
			if len(OnlineIds) == 0 {
				OnlineIds = append(OnlineIds, myElev.ID)

			}
			fmt.Println(OnlineIds)
			if len(OnlineIds) != 0 && len(p.Lost) != 0 {
				for j := 0; j < len(p.Lost); j++ {
					peerLost, _ := strconv.Atoi(p.Lost[j])
					for i := 0; i < len(OnlineIds); i++ {
						if OnlineIds[i] == peerLost {
							OnlineIds = append(OnlineIds[:i], OnlineIds[i+1:]...)
						}
					}

				}
			}
			fmt.Println(OnlineIds)
			if len(OnlineIds) != 0 && len(p.Lost) != 0 {
				for i := 0; i < len(AllElevators); i++ {
					for j := 0; j < len(p.Lost); j++ {
						peerLost, _ := strconv.Atoi(p.Lost[j])
						if AllElevators[i].ID == peerLost {
							ReassignOrders(AllElevators[i], myElev)
						}
					}
				}
			}
			if len(AllElevators) != 0 {
				fmt.Println(AllElevators)
				peerNew, _ := strconv.Atoi(p.New)
				if ContainsID(AllElevators, peerNew) {
					fmt.Println("inni i første if")
					for i := 0; i < len(AllElevators); i++ {
						for f := 0; f < NumFloors; f++ {
							if (AllElevators[i].ID == peerNew) && (AllElevators[i].Orders[f][eio.BT_Cab].Status == UT.Active) {
								AddElevToMsgQueue(AllElevators[i])
								fmt.Println("sender tilbake") //Kun ved hall orders
							}

						}
					}

				}
			}

		}
	}
}

func UpdateAllElevators() {
	if len(AllElevators) != 0 {
		for j := 0; j < len(AllElevators); j++ {
			if ListContains(OnlineIds, AllElevators[j].ID) {
				AllElevators[j].Online = true
			} else {
				AllElevators[j].Online = false
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

func ContainsID(list []UT.Elevator, new int) bool {
	for i := 0; i < len(list); i++ {
		if list[i].ID == new {
			return true
		}
	}
	return false
}

func AddHallOrderToMsgQueue(myElev UT.Elevator, btnFloor int, btnType eio.ButtonType) {
	iter++
	if len(AllElevators) == 0 {
		AllElevators = append(AllElevators, myElev)
	}

	weCanTakeIt := canTakeOrder()

	bestId := OD.CostCalculator(weCanTakeIt, btnFloor, btnType)
	for i := 0; i < len(AllElevators); i++ {
		AllElevators[i].Orders[btnFloor][btnType].Status = UT.Inactive
	}
	order := UT.Order{Floor: btnFloor, ButtonType: int(btnType), Status: UT.Active, Finished: false}
	Message.MsgID = iter
	Message.Elevator = myElev
	Message.IsNewOrder = true
	Message.Order = order
	Message.NewOrderTakerID = bestId
	Message.IsReceived = false
	Message.LocalID = myElev.ID
	MsgQueue = append(MsgQueue, Message)
}

func canTakeOrder() []UT.Elevator {
	weCanTakeIt := []UT.Elevator{}
	for i := 0; i < len(AllElevators); i++ {
		if (ListContains(OnlineIds, AllElevators[i].ID)) && !(AllElevators[i].MotorStop) {
			weCanTakeIt = append(weCanTakeIt, AllElevators[i])
		}
	}
	return weCanTakeIt
}

func UpdateHallLights() {
	for i := 0; i < len(AllElevators); i++ {
		for f := 0; f < UT.NumFloors; f++ {
			if AllElevators[i].Orders[f][eio.BT_HallUp].Status == UT.Active {
				eio.SetButtonLamp(eio.BT_HallUp, f, true)
			}

			if AllElevators[i].Orders[f][eio.BT_HallDown].Status == UT.Active {
				eio.SetButtonLamp(eio.BT_HallDown, f, true)
			}

		}
	}
	for f := 0; f < UT.NumFloors; f++ {
		for b := 0; b < 2; b++ {
			number := 0
			for i := 0; i < len(AllElevators); i++ {
				if AllElevators[i].Orders[f][b].Status == UT.Inactive {
					number++
				}
			}
			if len(AllElevators) == number {
				eio.SetButtonLamp(eio.ButtonType(b), f, false)
			}

		}
	}
	time.Sleep(10 * time.Millisecond)
}

func AddElevToMsgQueue(myElev UT.Elevator) {
	iter++
	Message.MsgID = iter
	Message.Elevator = myElev
	Message.IsNewOrder = false
	Message.IsReceived = false
	Message.LocalID = myElev.ID
	MsgQueue = append(MsgQueue, Message)
}

func SendMessage(msgChan UT.MsgChan, myElev UT.Elevator) {
	MsgTimeOut := time.NewTimer(200 * time.Millisecond)
	MsgTimeOut.Stop()

	for {
		if !(len(MsgQueue) == 0) {
			msg := MsgQueue[0]
			msgChan.SendChan <- msg
			MsgTimeOut.Reset(200 * time.Millisecond)
			if len(receivedMsg) >= (len(OnlineIds) - 1) {
				MsgQueue = MsgQueue[1:]
				receivedMsg = receivedMsg[:0]
				MsgTimeOut.Stop()
			} else {
				time.Sleep(10 * time.Millisecond)
			}

		}
		time.Sleep(4 * time.Millisecond) // Coopratrive routine
	}

	for {
		select {
		case <-MsgTimeOut.C:
			UpdateAllElevators()
			for i := 0; i < len(AllElevators); i++ {
				if !ListContains(receivedMsg, AllElevators[i].ID) {
					ReassignOrders(AllElevators[i], myElev)
					for f := 0; f < UT.NumFloors; f++ {
						for btn := 0; btn < UT.NumButtons-1; btn++ {
							myElev.Orders[f][btn].Status = UT.Inactive
						}
					}
				}
			}
		}
	}
}

func ReassignOrders(elev UT.Elevator, myElev UT.Elevator) {
	for i := 0; i < len(AllElevators); i++ {
		for f := 0; f < UT.NumFloors; f++ {
			for btn := 0; btn < UT.NumButtons-1; btn++ {
				if elev.Orders[f][btn].Status == UT.Active {
					AddHallOrderToMsgQueue(myElev, f, eio.ButtonType(btn))
					if AllElevators[i].ID == elev.ID {
						AllElevators[i].Orders[f][btn].Status = UT.Inactive
					}
				}
			}
		}
	}

}

func ConfirmationMessage(incomingMsg UT.Msg, myElev UT.Elevator, msgChan UT.MsgChan) {
	var ConMessage UT.Msg
	ConMessage.IsReceived = true
	ConMessage.IsNewOrder = false
	ConMessage.LocalID = myElev.ID
	ConMessage.MsgID = incomingMsg.MsgID
	msgChan.SendChan <- ConMessage
	time.Sleep(2 * time.Millisecond)
}

func Run(incomingMsg UT.Msg, myElev *UT.Elevator, msgChan UT.MsgChan) {
	if !(incomingMsg.LocalID == myElev.ID) {
		if incomingMsg.IsReceived {
			if !ListContains(receivedMsg, incomingMsg.LocalID) {
				receivedMsg = append(receivedMsg, incomingMsg.LocalID)
			}
		} else {
			ConfirmationMessage(incomingMsg, *myElev, msgChan)
		}
	}
	if !(incomingMsg.IsReceived) && !(incomingMsg.IsNewOrder) {
		if !(LastIncomingMessage.MsgID == incomingMsg.MsgID && LastIncomingMessage.LocalID == incomingMsg.LocalID) {
			LastIncomingMessage.MsgID = incomingMsg.MsgID
			LastIncomingMessage.LocalID = incomingMsg.LocalID
			if len(AllElevators) != 0 {
				if ContainsID(AllElevators, incomingMsg.LocalID) {
					for i := 0; i < len(AllElevators); i++ {
						if AllElevators[i].ID == incomingMsg.Elevator.ID {
							AllElevators[i] = incomingMsg.Elevator
						}
					}

				} else if !ContainsID(AllElevators, incomingMsg.LocalID) {
					AllElevators = append(AllElevators, incomingMsg.Elevator)
				}

			} else {
				//AllElevators = append(AllElevators, myElev)
			}
			if !ContainsID(AllElevators, incomingMsg.LocalID) {
				AllElevators = append(AllElevators, incomingMsg.Elevator)
				fmt.Println(AllElevators)
				for i := 0; i < len(AllElevators); i++ {
					if AllElevators[i].ID == myElev.ID {
						for f := 0; f < NumFloors; f++ {
							if AllElevators[i].Orders[f][eio.BT_Cab].Status == UT.Active {
								myElev.Orders[f][eio.BT_Cab].Status = UT.Active
							}

						}
					}
				}

			}
		}
	}
}

func ShouldITake(incomingMsg UT.Msg, myElev UT.Elevator) bool {
	shouldITake := false
	if incomingMsg.IsNewOrder && !incomingMsg.IsReceived {
		if !(LastIncomingMessage.MsgID == incomingMsg.MsgID && LastIncomingMessage.LocalID == incomingMsg.LocalID) {
			LastIncomingMessage.MsgID = incomingMsg.MsgID
			LastIncomingMessage.LocalID = incomingMsg.LocalID
			if !ContainsID(AllElevators, incomingMsg.LocalID) {
				AllElevators = append(AllElevators, incomingMsg.Elevator)
			}

			for i := 0; i < len(AllElevators); i++ {
				if AllElevators[i].ID == incomingMsg.Elevator.ID {
					AllElevators[i] = incomingMsg.Elevator
				}

				if incomingMsg.NewOrderTakerID == myElev.ID {
					shouldITake = true

				}
			}

		}
	}
	return shouldITake
}
