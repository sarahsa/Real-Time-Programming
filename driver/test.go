package main


import(
	"../network/network/bcast"
	"../network/network/peers"
	"fmt"
	"flag"
	"../elevio"
	//"../OrderManager"
	"../config"
	"../Fsm"

)

//Denne burde vaare i nettverkmodulen
type ButtonPressPacket struct{
	Sender string
	//BT elevio.ButtonEvent
	Floor int
	Button elevio.ButtonType
}

/*type ButtonPressPacket struct{
	Sender string
	currentFloor int
	//currentDirection int
	//currentState int //???
	ButtonPressed elevio.ButtonEvent
}*/



func main() {
	//Fungerte ved aa fa port og id fra terminalen.
	idPtr := flag.String("id", "no id", "elevatro id")
	hwPortPtr := flag.String("hwport", "15657", "select w port hw runs on")
	flag.Parse()
	myID := *idPtr
	hwPort := *hwPortPtr

	/*localIP, err := localip.localIP()
	if err != nil{
		fmt.Println(err)
		fmt.Println("Diconnected")
		os.Exit(0)
	}
	id := fmt.Sprintf("peer-%s", localIP)*/


	elevio.Init(":"+hwPort, 4)

	Ch_assignedOrders := make(chan elevio.ButtonEvent)
	ch_floors  := make(chan int)
    ch_doorTimeout := make(chan bool)



	ButtonPacketTrans := make(chan ButtonPressPacket)
	ButtonPacketRecv := make(chan ButtonPressPacket)

	ElevatorTrans := make(chan config.Elevator)
	ElevatorRecv := make(chan config.Elevator)

	ButtonPress := make(chan elevio.ButtonEvent)
	//Ch_elevator := make(chan config.Elevator)
	//Ch_direction := make(chan elevio.MotorDirection)

	//go Fsm.UpdateElevator(Ch_elvator)
	peerUpdateCh := make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool)

	go peers.Transmitter(15647, myID, peerTxEnable)
	go peers.Receiver(15647, peerUpdateCh)

	go bcast.Transmitter(23232, ButtonPacketTrans, ElevatorTrans)
	go bcast.Receiver(23232, ButtonPacketRecv, ElevatorRecv)


	go elevio.PollButtons(ButtonPress)
    go elevio.PollFloorSensor(ch_floors)

	go Fsm.Fsm(Ch_assignedOrders, ch_floors, ch_doorTimeout)


	for{
		select{
		case p := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)
			


		case buttonPress := <-ButtonPress:
			fmt.Println("Button press at " + myID)
			fmt.Println(buttonPress)

			//floor <- Ch_floor

			//OrderManager.executeOrder(buttonPress)
			//fmt.Println("Change Made: ", changeMade)

			ButtonPacketTrans <- ButtonPressPacket{myID, buttonPress.Floor, buttonPress.Button}
			//ElevatorTrans <- Ch_elevator



		case recvPacket := <-ButtonPacketRecv:
			Ch_assignedOrders <- elevio.ButtonEvent{recvPacket.Floor, recvPacket.Button}

			fmt.Println("Received from " + recvPacket.Sender)
			fmt.Println(recvPacket.Floor, " ", recvPacket.Button)

		/*	orderAccepted, changeMade := OrderManager.AddOrder(buttonPress)
			fmt.Println("Change Made: ", changeMade)
			if changeMade {
				ButtonPacketRecv <- ButtonPressPacket{myID, buttonPress.Floor, int(buttonPress.Button)}
			}*/
/*
		case recvElevPacket := <- ElevatorRecv:

			for k, v := range OrderManager.elevators {
				if k != recvPacket.ID {
					elevators[recvPacket.ID] = recvPacket
				}
			}*/

		}

	}


}
