package main

import (
	//"../network/network/bcast"
	//"../network/network/peers"
	"flag"
	"fmt"
	"os"

	"../Fsm"
	"../OrderManager"
	"../config"
	"../elevio"
	"../network/network/localip"
	//"../sync"
)

func main() {

	//Fungerte ved aa fa port og id fra terminalen.
	/*idPtr := flag.String("id", "no id", "elevatro id")
	//flag.StringVar(&myID, "id", "", "id of this peer")
	hwPortPtr := flag.String("hwport", "15657", "select w port hw runs on")
	flag.Parse()
	myID := *idPtr
	hwPort := *hwPortPtr*/

	var myID string
	flag.StringVar(&myID, "id", "", "id of this peer")
	hwPortPtr := flag.String("hwport", "15657", "select w port hw runs on")
	flag.Parse()
	hwPort := *hwPortPtr

	elevio.Init(":"+hwPort, 4)

	if myID == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			fmt.Println("Diconnected")
			localIP = "DISCONNECTED"
		}
		myID = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
	}

	//elevio.Init(":"+hwPort, 4) //4 = number of floors

	assignedOrders := make(chan elevio.ButtonEvent, 10)
	doorTimeout := make(chan bool)

	UpdatedElevatorStatus := make(chan config.Elevator, 10)

	//These four channels might be moved to OM
	ButtonPacketTrans := make(chan config.OrderPacket, 10)
	ButtonPacketRecv := make(chan config.OrderPacket, 10)

	ElevatorPacketTrans := make(chan config.ElevatorStatusPacket, 10)
	ElevatorPacketRecv := make(chan config.ElevatorStatusPacket, 10)

	AckReceivedOrderTrans := make(chan config.AcknowledgmentPacket, 10)
	AckReceivedOrderRecv := make(chan config.AcknowledgmentPacket, 10)

	AckExecutedOrderTrans := make(chan elevio.ButtonEvent, 10)
	AckExecutedOrderRecv := make(chan elevio.ButtonEvent, 10)

	ButtonPress := make(chan elevio.ButtonEvent, 10)
	OrderIsExecuted := make(chan elevio.ButtonEvent, 10)

	MotorTimedOut := make(chan config.OrderMatrix, 10)

	// go Fsm.UpdateElevator(Ch_elvator)

	go OrderManager.OrderManager(ButtonPacketTrans, ButtonPacketRecv, assignedOrders,
		doorTimeout, ElevatorPacketTrans, ElevatorPacketRecv, ButtonPress, myID,
		hwPort, UpdatedElevatorStatus, AckReceivedOrderRecv, AckReceivedOrderTrans, OrderIsExecuted, AckExecutedOrderRecv, AckExecutedOrderTrans, MotorTimedOut)
	go Fsm.Fsm(assignedOrders, doorTimeout, UpdatedElevatorStatus, OrderIsExecuted, MotorTimedOut)

	//fmt.Println("elevator.State: %d", elevator.State)
	//go Fsm.Fsm(assignedOrders, floors, doorTimeout)

	select {}

}

//-----------------------------------------------------------------------------------------------------

/*
package main

import(
	"../network/network/bcast"
	"../network/network/peers"
	//"../network/network/localip"
	"fmt"
	"flag"
	"../elevio"
//"../OrderManager"
	"../config"
	//"../Fsm"

)
func main() {

	//Fungerte ved aa fa port og id fra terminalen.
	/*idPtr := flag.String("id", "no id", "elevatro id")
	//flag.StringVar(&myID, "id", "", "id of this peer")
	hwPortPtr := flag.String("hwport", "15657", "select w port hw runs on")
	flag.Parse()
	myID := *idPtr
	hwPort := *hwPortPtr

	var myID string
	flag.StringVar(&myID, "id", "", "id of this peer")
	hwPortPtr := flag.String("hwport", "15657", "select w port hw runs on")
	flag.Parse()
	hwPort := *hwPortPtr

	/*var myID string
	if myID == ""{
		localIP, err := localip.LocalIP()
		if err != nil{
			fmt.Println(err)
			fmt.Println("Diconnected")
			localIP = "DISCONNECTED"
		}
		myID = fmt.Sprintf("peer-%s", localIP)
	}


	elevio.Init(":"+hwPort, 4) //4 = number of floors

	assignedOrders := make(chan elevio.ButtonEvent)
	floors := make(chan int)
    //doorTimeout := make(chan bool)

	ButtonPacketTrans := make(chan config.ButtonPressPacket)
	ButtonPacketRecv := make(chan config.ButtonPressPacket)

	ElevatorTrans := make(chan config.Elevator)
	ElevatorRecv := make(chan config.Elevator)

	ButtonPress := make(chan elevio.ButtonEvent)

	// go Fsm.UpdateElevator(Ch_elvator)
	peerUpdateCh := make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool)
	go peers.Transmitter(15647, myID, peerTxEnable) //15647 , 15670
	go peers.Receiver(15647, peerUpdateCh) //15647

	go bcast.Transmitter(23232, ButtonPacketTrans, ElevatorTrans)
	go bcast.Receiver(23232, ButtonPacketRecv, ElevatorRecv)

	go elevio.PollButtons(ButtonPress)
	go elevio.PollFloorSensor(floors)

	//go Fsm.Fsm(assignedOrders, floors, doorTimeout)

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

				ButtonPacketTrans <- config.ButtonPressPacket{myID, buttonPress.Floor, buttonPress.Button}

			case recvPacket := <-ButtonPacketRecv:
			assignedOrders <- elevio.ButtonEvent{recvPacket.Floor, recvPacket.Button}

			fmt.Println("Received from " + recvPacket.Sender)
			fmt.Println(recvPacket.Floor, " ", recvPacket.Button)

		}

	}
	 select{}

}

*/
