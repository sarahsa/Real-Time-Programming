package main

import (
	"flag"
	"fmt"
	"os"
	"../Fsm"
	"../OrderManager"
	"../config"
	"../elevio"
	"../network/network/localip"
)

func main() {

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

	ch_assignedOrders := make(chan elevio.ButtonEvent, 10)
	ch_doorTimeout := make(chan bool)

	ch_updatedElevatorStatus := make(chan config.Elevator, 10)

	ch_buttonPacketTrans := make(chan config.OrderPacket, 10)
	ch_buttonPacketRecv := make(chan config.OrderPacket, 10)

	ch_elevatorPacketTrans := make(chan config.ElevatorStatusPacket, 10)
	ch_elevatorPacketRecv := make(chan config.ElevatorStatusPacket, 10)

	ch_ackReceivedOrderTrans := make(chan config.AcknowledgmentPacket, 10)
	ch_ackReceivedOrderRecv := make(chan config.AcknowledgmentPacket, 10)

	ch_ackExecutedOrderTrans := make(chan elevio.ButtonEvent, 10)
	ch_ackExecutedOrderRecv := make(chan elevio.ButtonEvent, 10)

	ch_buttonPress := make(chan elevio.ButtonEvent, 10)
	ch_orderIsExecuted := make(chan elevio.ButtonEvent, 10)

	ch_timedOutMotor := make(chan config.OrderMatrix, 10)

	go OrderManager.OrderManager(ch_buttonPacketTrans, ch_buttonPacketRecv, ch_assignedOrders,
		ch_doorTimeout, ch_elevatorPacketTrans, ch_elevatorPacketRecv, ch_buttonPress, myID,
		hwPort, ch_updatedElevatorStatus, ch_ackReceivedOrderRecv, ch_ackReceivedOrderTrans,
		ch_orderIsExecuted, ch_ackExecutedOrderRecv, ch_ackExecutedOrderTrans, ch_timedOutMotor)

	go Fsm.Fsm(ch_assignedOrders, ch_doorTimeout, ch_updatedElevatorStatus, ch_orderIsExecuted, ch_timedOutMotor)

	select {}

}
