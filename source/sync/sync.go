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
var OnlineIds = []int{}

func UpdateOnlineIds(peerUpdateCh chan peers.PeerUpdate, myElev UtilitiesTypes.Elevator) {
	for {
		select {
		case p := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
				fmt.Printf("  Peers:    %q\n", p.Peers)
				fmt.Printf("  New:      %q\n", p.New)
				fmt.Printf("  Lost:     %q\n", p.Lost)
				for i := 0; i < len(p.Peers); i++ {
					peer, _ := strconv.Atoi(p.Peers[i])
				if len(OnlineIds)!=0{
						if !ListContains(OnlineIds, peer) {
							OnlineIds = append(OnlineIds, peer)

					}
				}
			}
			if len(OnlineIds)==0{
				OnlineIds = append(OnlineIds, myElev.ID)

			}
			fmt.Println(OnlineIds)
					if len(OnlineIds)!=0 && len(p.Lost)!=0{
						for j := 0; j < len(p.Lost); j++ {
							peerLost, _ := strconv.Atoi(p.Peers[j])
							for i := 0; i < len(OnlineIds); i++ {
								if OnlineIds[i] == peerLost {
									OnlineIds = append(OnlineIds[:i],OnlineIds[i+1:]...)
								}
							}
							

				}
			}
			fmt.Println(OnlineIds)
			
		
	}
	}
}

func UpdateOnlineElevators(){
	if len(OnlineElevators) !=0 {
		for j := 0; j < len(OnlineElevators); j++ {
			if ListContains(OnlineIds, OnlineElevators[j].ID){
				OnlineElevators[j].Online = true
				fmt.Println(OnlineElevators[j].ID, "online")
		} else {
			OnlineElevators[j].Online = false
			fmt.Println(OnlineElevators[j].ID, "offline")

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
	if len(OnlineElevators) == 0 {
		OnlineElevators = append(OnlineElevators,myElev)
	}
	weCanTakeIt:= canTakeOrder()
	bestId := OrderDistributor.CostCalculator(weCanTakeIt, btnFloor, btnType)
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

func canTakeOrder() []UtilitiesTypes.Elevator{
	weCanTakeIt := [] UtilitiesTypes.Elevator{}
	for i:= 0 ; i < len(OnlineElevators); i++{
		if (ListContains(OnlineIds,OnlineElevators[i].ID )) && !(OnlineElevators[i].MotorStop){
			weCanTakeIt = append(weCanTakeIt, OnlineElevators[i])
		} 
	}
return weCanTakeIt
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

func Run(incomingMsg UtilitiesTypes.Msg,myElev UtilitiesTypes.Elevator, msgChan UtilitiesTypes.MsgChan) {
	UpdateHallLights()
	
	//msgTimeout := time.NewTimer(20 * time.Millisecond)
	if !(incomingMsg.LocalID == myElev.ID) {
		if incomingMsg.IsReceived {
			if !ListContains(receivedMsg, incomingMsg.LocalID) {
				receivedMsg = append(receivedMsg, incomingMsg.LocalID)
				if len(receivedMsg) >= numPeers {
					//msgTimeout.Stop()

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
						if OnlineElevators[i].ID == incomingMsg.Elevator.ID {
							OnlineElevators[i] = incomingMsg.Elevator
						}
					}
				} else if !ContainsID(OnlineElevators, incomingMsg.LocalID) {
					if incomingMsg.LocalID != 0 {
						OnlineElevators = append(OnlineElevators, incomingMsg.Elevator)
					}
				}
			} else {
				OnlineElevators = append(OnlineElevators, myElev)
			}

		}
	}
	/*
case <-msgTimeout.C:
	UpdateOnlineElevators()
	for i:=0; i < len(OnlineElevators); i++ {
		if !ListContains(receivedMsg, OnlineElevators[i].ID){
			for f :=0; f < UtilitiesTypes.NumFloors; f++{
				for btn:=0; btn < UtilitiesTypes.NumButtons-1; btn++{
					if OnlineElevators[i].Orders[f][btn].Status == UtilitiesTypes.Active{
					AddHallOrderToMsgQueue(OnlineElevators[i],f, elevio.ButtonType(btn))
					}
				}
			}*/

	
	
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
