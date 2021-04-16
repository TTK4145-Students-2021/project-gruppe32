package sync

import (
	"fmt"
	"time"

	"../OrderDistributor"
	"../UtilitiesTypes"
	"../elevio"
)

/*
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

}*/
var (
 MsgQueue = []UtilitiesTypes.Msg{}
 onlineIDs []int
 OnlineElevators []UtilitiesTypes.Elevator
 allElevators [UtilitiesTypes.NumElevs]UtilitiesTypes.Elevator
 iter = 0
 Message UtilitiesTypes.Msg
 elev UtilitiesTypes.Elevator


	numPeers = 1
	receivedMsg         []int
	LastIncomingMessage UtilitiesTypes.Msg
)

/*
func TestingNetworkElev() {
	for i := 0; i < len(OnlineElevators); i++ {
		fmt.Println(OnlineElevators[i].Orders)
	}
}*/

func Sync(msgChan UtilitiesTypes.MsgChan, fsmChan UtilitiesTypes.FsmChan, id int){
	go func() {
		for {
			select{
			case elev = <-fsmChan.Elev:
				allElevators[id] = elev
			}
		}
	}()
	msgTimer := time.NewTimer(10 * time.Second)
	msgTimer.Stop()
	for {
		select{
		case incomingMsg := <- msgChan.RecChan:
			recID := incomingMsg.LocalID
			fmt.Println("recID: ", recID)
			fmt.Println("min id", id)
			if id != recID{
				if !ListContains(onlineIDs, recID){
				onlineIDs = append(onlineIDs, recID)
				numPeers = len(onlineIDs)
				}
				if incomingMsg.IsReceived{
					if incomingMsg.MsgID == LastIncomingMessage.MsgID {
						if !ListContains(receivedMsg, recID){
							receivedMsg = append(receivedMsg,recID)
							if len(receivedMsg) == numPeers{
								//msgTimer.Stop()
								
							}
						}
					}
				}else{
					allElevators[recID] = incomingMsg.Elevator
					for e := 0; e < UtilitiesTypes.NumElevs; e++ {
						if (!ListContains(onlineIDs, allElevators[e].ID)) && (e!=id){
							allElevators[e].Orders = [UtilitiesTypes.NumFloors][UtilitiesTypes.NumButtons]UtilitiesTypes.Order{}
						}
					}
					ConfirmationMessage(incomingMsg,elev,msgChan)
					fmt.Println("sendCon")

				}
				if incomingMsg.IsNewOrder{
					if incomingMsg.Order.OrderTaker == id{
						elev.Orders[incomingMsg.Order.Floor][incomingMsg.Order.ButtonType].Status = UtilitiesTypes.Active
						new := elevio.ButtonEvent{Floor: incomingMsg.Order.Floor, Button: elevio.ButtonType(incomingMsg.Order.ButtonType)}
						fsmChan.NewOrder <- new
					}
			
				}
			}
			fsmChan.Elev <- elev

			LastIncomingMessage.MsgID = incomingMsg.MsgID
			LastIncomingMessage.LocalID = incomingMsg.LocalID

			//case :
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



func AddHallOrderToMsgQueue(myElev UtilitiesTypes.Elevator, btnFloor int, btnType elevio.ButtonType) {
	iter++
	bestId := OrderDistributor.CostCalculator(allElevators, btnFloor, btnType)
	/*if bestId == myElev.ID {
		myElev.Orders[btnFloor][btnType].Status = UtilitiesTypes.Active
		AddElevToMsgQueue(*myElev)
	}*/
	order := UtilitiesTypes.Order{OrderTaker: bestId,Floor: btnFloor, ButtonType: int(btnType), Status: UtilitiesTypes.Inactive, Finished: false}
	Message.MsgID = iter
	Message.Elevator = myElev
	Message.IsNewOrder = true
	Message.Order = order
	Message.IsReceived = false
	Message.LocalID = myElev.ID
	MsgQueue = append(MsgQueue, Message)
	//fmt.Println(Message.Elevator.Orders)
	//fmt.Println("hall order")
	//fmt.Println(bestId)

}

func UpdateHallLights() {

	for i := 0; i < len(allElevators); i++ {
		for f := 0; f < UtilitiesTypes.NumFloors; f++ {
			if allElevators[i].Orders[f][elevio.BT_HallUp].Status == UtilitiesTypes.Active {
				elevio.SetButtonLamp(elevio.BT_HallUp, f, true)
			}

			if allElevators[i].Orders[f][elevio.BT_HallDown].Status == UtilitiesTypes.Active {
				elevio.SetButtonLamp(elevio.BT_HallDown, f, true)
			}

		}
	}
	for f := 0; f < UtilitiesTypes.NumFloors; f++ {
		for b := 0; b < 2; b++ {
			number := 0
			for i := 0; i < len(allElevators); i++ {
				if allElevators[i].Orders[f][b].Status == UtilitiesTypes.Inactive {
					number++
				}
			}
			if len(allElevators) == number {
				elevio.SetButtonLamp(elevio.ButtonType(b), f, false)
			}

		}
	}
	time.Sleep(10 * time.Millisecond)

}

func ClearHallAtCurrentFloor(myElev UtilitiesTypes.Elevator) {
	for btn := 0; btn < 2; btn++ {
		for i := 0; i < len(allElevators); i++ {
			allElevators[i].Orders[myElev.Floor][btn].Status = UtilitiesTypes.Inactive
			allElevators[i].Orders[myElev.Floor][btn].Finished = true
		}
	}
}

func AddElevToMsgQueue(fsmChan UtilitiesTypes.FsmChan) {
	for {
		select{

		case elev := <- fsmChan.Elev:
		iter++
		Message.MsgID = iter
		Message.Elevator = elev
		Message.IsNewOrder = false
		Message.IsReceived = false
		//fmt.Println("elevid", elev.ID)
		Message.LocalID = elev.ID
		//fmt.Println("messageID", Message.LocalID)
		MsgQueue = append(MsgQueue, Message)
		}
	}
}

func SendMessage(msgChan UtilitiesTypes.MsgChan) {
	for {
		if !(len(MsgQueue) == 0) {
			msg := MsgQueue[0]
			msgChan.SendChan <- msg
			if len(receivedMsg) >= numPeers {
				fmt.Println("if")
				MsgQueue = MsgQueue[1:]
				receivedMsg = receivedMsg[:0]
			} else {
				
				
			}

		}
		time.Sleep(10 * time.Millisecond) // Coopratrive routine
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
	ConMessage.Order = UtilitiesTypes.Order{OrderTaker:-1,Floor:-1,ButtonType:-1,Status:-1,Finished:false}
	msgChan.SendChan <- ConMessage
	time.Sleep(2 * time.Millisecond)
}

func Run(fsmChan UtilitiesTypes.FsmChan, msgChan UtilitiesTypes.MsgChan) {
	myElev := <- fsmChan.Elev
	//UpdateHallLights()
	//fmt.Println(incomingMsg.Elevator.ID)
	//for i := 0; i < len(OnlineElevators); i++ {
	//fmt.Println(OnlineElevators[i].ID)
	//fmt.Println("---------", i)
	//}
	for{
	select{
	case incomingMsg := <- msgChan.RecChan:
	if !(incomingMsg.LocalID == myElev.ID) {
		//fmt.Println(incomingMsg.LocalID)
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
	if !(incomingMsg.IsReceived) {
	if !(incomingMsg.IsNewOrder) {

		if !(LastIncomingMessage.MsgID == incomingMsg.MsgID && LastIncomingMessage.LocalID == incomingMsg.LocalID) {
			LastIncomingMessage.MsgID = incomingMsg.MsgID
			LastIncomingMessage.LocalID = incomingMsg.LocalID
			if len(OnlineElevators) != 0 {
				if ContainsID(OnlineElevators, incomingMsg.LocalID) {
					for i := 0; i < len(allElevators); i++ {
						//fmt.Println(OnlineElevators[i].ID, "Run")
						//fmt.Println(OnlineElevators[i].Orders, "\n")
						if allElevators[i].ID == incomingMsg.Elevator.ID {
							allElevators[i] = incomingMsg.Elevator
							//fmt.Println("etter, Run")
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
	if incomingMsg.IsNewOrder{
		if incomingMsg.Order.OrderTaker == myElev.ID{
			elev.Orders[incomingMsg.Order.Floor][incomingMsg.Order.ButtonType].Status = UtilitiesTypes.Active
			new := elevio.ButtonEvent{Floor: incomingMsg.Order.Floor, Button: elevio.ButtonType(incomingMsg.Order.ButtonType)}
			fsmChan.NewOrder <- new
		}

	}
	//fmt.Println(OnlineElevators)
	//fmt.Println("her kommer de på nytt")
}
}
	}
}
/*
func ShouldITake(incomingMsg UtilitiesTypes.Msg, myElev UtilitiesTypes.Elevator) bool {
	shouldITake := false
	if incomingMsg.IsNewOrder && !incomingMsg.IsReceived {
		/*
		if !(LastIncomingMessage.MsgID == incomingMsg.MsgID && LastIncomingMessage.LocalID == incomingMsg.LocalID) {
			LastIncomingMessage.MsgID = incomingMsg.MsgID
			LastIncomingMessage.LocalID = incomingMsg.LocalID
			if !ContainsID(OnlineElevators, incomingMsg.LocalID) {
				OnlineElevators = append(OnlineElevators, incomingMsg.Elevator)
			}

			for i := 0; i < len(OnlineElevators); i++ {
				fmt.Println(OnlineElevators[i].ID, "ShouldITake")
				fmt.Println(OnlineElevators[i].Orders, "\n")
				if OnlineElevators[i].ID == incomingMsg.Elevator.ID {
					OnlineElevators[i] = incomingMsg.Elevator
					fmt.Println("etter ShouldITake")
					fmt.Println(OnlineElevators[i].ID)
					fmt.Println(OnlineElevators[i].Orders, "\n")

				}
			}

			if incomingMsg.NewOrderTakerID == myElev.ID {
				shouldITake = true

			}
		}

	//}
	//fmt.Println(OnlineElevators)
	//fmt.Println("her kommer de på nytt")
	return shouldITake
}

/*
func Sync(msgChan UtilitiesTypes.MsgChan, myElev *UtilitiesTypes.Elevator) {
	for {
		select {
		case incomingMsg := <-msgChan.RecChan:
			if !(incomingMsg.LocalID == myElev.ID) {

				if incomingMsg.IsReceived {
					fmt.Println("Is Recived")
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
					ConfirmationMessage(incomingMsg, *myElev, msgChan)
					fmt.Println("Con Message")
					if !(LastIncomingMessage.MsgID == incomingMsg.MsgID && LastIncomingMessage.LocalID == incomingMsg.LocalID) {
						if !ContainsID(OnlineElevators, incomingMsg.LocalID) {
							OnlineElevators = append(OnlineElevators, incomingMsg.Elevator)
						}
						for i := 0; i < len(OnlineElevators); i++ {
							if OnlineElevators[i].ID == incomingMsg.LocalID {
								OnlineElevators[i] = incomingMsg.Elevator
							}
						}
					}
				}
			}
			if incomingMsg.IsNewOrder {

				if !(LastIncomingMessage.MsgID == incomingMsg.MsgID && LastIncomingMessage.LocalID == incomingMsg.LocalID) {
					LastIncomingMessage.MsgID = incomingMsg.MsgID
					LastIncomingMessage.LocalID = incomingMsg.LocalID
					if !ContainsID(OnlineElevators, incomingMsg.LocalID) {
						OnlineElevators = append(OnlineElevators, incomingMsg.Elevator)
					}
					for i := 0; i < len(OnlineElevators); i++ {
						if OnlineElevators[i].ID == incomingMsg.LocalID {
							OnlineElevators[i] = incomingMsg.Elevator
							fmt.Println(OnlineElevators[i].Floor)
						}
						if incomingMsg.NewOrderTakerID == myElev.ID {
							myElev.Orders[incomingMsg.Order.Floor][incomingMsg.Order.ButtonType].Status = UtilitiesTypes.Active
							fmt.Println(myElev.Orders)
						}

					}
				}
			}
		}
	}
}



func Networkmain() {
	// Our id can be anything. Here we pass it on the command line, using
	//  `go run main.go -id=our_id`
	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()

	// ... or alternatively, we can use the local IP address.
	// (But since we can run multiple programs on the same PC, we also append the
	//  process ID)
	if id == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		id = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
	}

	// We make a channel for receiving updates on the id's of the peers that are
	//  alive on the network
	peerUpdateCh := make(chan peers.PeerUpdate)
	// We can disable/enable the transmitter after it has been started.
	// This could be used to signal that we are somehow "unavailable".
	peerTxEnable := make(chan bool)
	go peers.Transmitter(15647, id, peerTxEnable)
	go peers.Receiver(15647, peerUpdateCh)

	// We make channels for sending and receiving our custom data types
	helloTx := make(chan HelloMsg)
	helloRx := make(chan HelloMsg)
	// ... and start the transmitter/receiver pair on some port
	// These functions can take any number of channels! It is also possible to
	//  start multiple transmitters/receivers on the same port.
	go bcast.Transmitter(16569, helloTx)
	go bcast.Receiver(16569, helloRx)

	// The example message. We just send one of these every second.
	go func() {
		helloMsg := HelloMsg{"Hello from " + id, 0}
		for {
			helloMsg.Iter++
			helloTx <- helloMsg
			time.Sleep(1 * time.Second)
		}
	}()

	fmt.Println("Started")
	for {
		select {
		case p := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)

		case a := <-helloRx:
			fmt.Printf("Received: %#v\n", a)
		}
	}
}

*/
