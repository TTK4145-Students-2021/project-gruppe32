package sync

import (
	"fmt"
	"time"

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
	bestId := OrderDistributor.CostCalculator(Heis3, OtherElevators, 1, 0)
	fmt.Println(bestId)

}

var buffer = make(chan UtilitiesTypes.Msg, 100)
var OtherElevators = []UtilitiesTypes.Elevator{}

func ListContains(list []int, new int) bool{
	for i := 0; i < len(list); i++ {
		if list[i] == new{
			return true
		}
	}
	return false
}

func ContainsID(list []UtilitiesTypes.Elevator, new int) bool{
	for i := 0; i < len(list); i++ {
		if list[i].ID == new{
			return true
		}
	}
	return false
}

var iter = 0
var Message UtilitiesTypes.Msg

var (
	numPeers int
	receivedMsg []int
	LastIncomingMessage UtilitiesTypes.Msg
)
	




func AddToMsgBuf(myElev UtilitiesTypes.Elevator, order UtilitiesTypes.Order, id int, newOrder bool) {
	iter ++
	Message.MsgID =iter
	Message.Elevator = myElev
	Message.IsNewOrder = newOrder
	Message.NewOrderTakerID = id
	Message.IsReceived = false
	Message.LocalID = myElev.ID

	buffer <- Message
}

func SendMessage(msgChan UtilitiesTypes.MsgChan){
	msg := <-buffer
	for {
		fmt.Println(msg)
		msgChan.SendChan <- msg
		if len(receivedMsg) >= numPeers{
			msg := <- buffer
			receivedMsg = receivedMsg[:0]
			msgChan.SendChan <- msg
		}
		if msg == 0 {
			break
		}
	// hvis timeren har gått ut, og vi må regne ut kostfunksjon på nytt hvis vi ikke får bekreftelse fra heisen som skulle ta ordren
	// Alle hall orders til heisen som ikke lenger er i Peers må noen andre heiser ta ordrene.
}
	
}

func ConfirmationMessage(incomingMsg UtilitiesTypes.Msg, myElev UtilitiesTypes.Elevator, msgChan UtilitiesTypes.MsgChan){
	var ConMessage UtilitiesTypes.Msg
	ConMessage.IsReceived = true
	ConMessage.LocalID = myElev.ID
	ConMessage.MsgID = incomingMsg.MsgID
	msgChan.SendChan <- ConMessage
	time.Sleep(2*time.Millisecond)
	


}

func Sync(msgChan UtilitiesTypes.MsgChan, myElev UtilitiesTypes.Elevator) {
	for{
		select{
		case incomingMsg := <- msgChan.RecChan:
			if !(LastIncomingMessage.MsgID==incomingMsg.MsgID && LastIncomingMessage.LocalID==incomingMsg.LocalID){

			if incomingMsg.IsReceived{
					if !ListContains(receivedMsg, incomingMsg.LocalID){
						receivedMsg = append(receivedMsg,incomingMsg.LocalID)
						if len(receivedMsg) >= numPeers{
							// stoppe timer??
							
						}
					}
					// hvis timeren har gått ut, og vi må regne ut kostfunksjon på nytt hvis vi ikke får bekreftelse fra heisen som skulle ta ordren
				// Alle hall orders til heisen som ikke lenger er i Peers må noen andre heiser ta ordrene.
				// 
			} else if incomingMsg.IsNewOrder{
				ConfirmationMessage(incomingMsg, myElev, msgChan)
				fmt.Println(incomingMsg.Elevator)
				fmt.Println(incomingMsg.Order)
				if incomingMsg.NewOrderTakerID == myElev.ID {
					myElev.Orders[incomingMsg.Order.Floor][incomingMsg.Order.ButtonType].Status = UtilitiesTypes.Active
				}

			}
			} else{
				ConfirmationMessage(incomingMsg, myElev, msgChan)
				if !ContainsID(OtherElevators, incomingMsg.LocalID){
					OtherElevators = append(OtherElevators, incomingMsg.Elevator)
				}
				for i:=0; i < len(OtherElevators);i++ {
					if OtherElevators[i].ID == incomingMsg.LocalID {
						OtherElevators[i] = incomingMsg.Elevator
				}
			}
		
		}
	}
}


}
	
	


/*

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