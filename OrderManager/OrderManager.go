package OrderManager

import (
	"fmt"
	"math"
	"../config"
	"../elevio"
	"../network/network/bcast"
	"../network/network/peers"
	"time"
	"strconv"
	"../Fsm"
	"../sync"
	"../backUp"
)


var allUpdatedElevators = make(map[string]config.Elevator)

var activeElevators = make(map[string]config.Elevator)

var cabOrders[config.N_FLOORS]bool

var orderRegistered [4][2]bool

func OrderManager(ch_newOrderTrans chan config.OrderPacket,
	ch_newOrderRecv chan config.OrderPacket,
	ch_assignedOrders chan elevio.ButtonEvent,
	ch_doorTimeout chan bool,
	ch_elevatorPacketTrans chan config.ElevatorStatusPacket,
	ch_elevatorPacketRecv chan config.ElevatorStatusPacket,
	ch_buttonPress chan elevio.ButtonEvent,
	myID string,
	hwPort string,
	ch_updatedElevatorStatus chan config.Elevator,
	ch_ackReceivedOrderRecv chan config.AcknowledgmentPacket,
	ch_ackReceivedOrderTrans chan config.AcknowledgmentPacket,
	ch_orderIsExecuted chan elevio.ButtonEvent,
	ch_ackExecutedRecv chan elevio.ButtonEvent,
	ch_ackExecutedTrans chan elevio.ButtonEvent,
	ch_motorTimedOut chan config.OrderMatrix) {

	ch_peerUpdate := make(chan peers.PeerUpdate)
	ch_peerTxEnable := make(chan bool)

	go peers.Transmitter(15847, myID, ch_peerTxEnable)
	go peers.Receiver(15847, ch_peerUpdate)

	go bcast.Transmitter(23232, ch_newOrderTrans, ch_elevatorPacketTrans, ch_ackReceivedOrderTrans, ch_ackExecutedTrans)
	go bcast.Receiver(23232, ch_newOrderRecv, ch_elevatorPacketRecv, ch_ackReceivedOrderRecv, ch_ackExecutedRecv)

	go elevio.PollButtons(ch_buttonPress)

	go sync.SendElevatorUpdate(allUpdatedElevators[myID],
		ch_updatedElevatorStatus,
		ch_elevatorPacketTrans,
		myID)

	for {
		select {
		case p := <-ch_peerUpdate:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)

			if p.New != "" {
				AddElevator(p.New, Fsm.GetElevatorStatus())
				activeElevators[p.New] = allUpdatedElevators[p.New]
			}

			if len(p.Lost) ==1 && len(p.Peers) > 1 {
				for i := range p.Lost {
					for f := 0; f < config.N_FLOORS; f++ {
						for b := 0; b < config.N_BUTTONS; b++ {
							if allUpdatedElevators[p.Lost[i]].AssignedRequests[f][b] == true{
								ch_buttonPress <- elevio.ButtonEvent{f, elevio.ButtonType(b)}

							}
						}
					}
				delete(allUpdatedElevators, p.Lost[i])
				}
			}

		case buttonPress := <-ch_buttonPress:
			if buttonPress.Button == elevio.BT_Cab {
				cabOrders[buttonPress.Floor] = true
				backUp.SaveToDisk(buttonPress, cabOrders)
				ch_assignedOrders <- buttonPress

			} else if len(activeElevators) > 1 {
				executer := CalculateCost(buttonPress)
				ch_newOrderTrans <- config.OrderPacket{executer, buttonPress}

			}
		case recvButtonPacket := <-ch_newOrderRecv:
			ch_ackReceivedOrderTrans <- config.AcknowledgmentPacket{myID, recvButtonPacket.Executer, recvButtonPacket.Button}

		case ack := <-ch_ackReceivedOrderRecv:
			if ack.Sender != myID {
				orderRegistered[(ack.Button).Floor][int((ack.Button).Button)] = true
				elevio.SetButtonLamp((ack.Button).Button, (ack.Button).Floor, true)
			}

			if ack.Executer == myID {
				ch_assignedOrders <- ack.Button
			}

		case order := <-ch_orderIsExecuted:
			cabOrders[order.Floor] = false
			backUp.SaveToDisk(order, cabOrders)

			if order.Button != elevio.BT_Cab {
				ch_ackExecutedTrans <- order
			}

		case ackExecuted := <-ch_ackExecutedRecv:
			elevio.SetButtonLamp(ackExecuted.Button, ackExecuted.Floor, false)
			orderRegistered[ackExecuted.Floor][int(ackExecuted.Button)] = false

		case updatedElevator := <-ch_elevatorPacketRecv:
			allUpdatedElevators[updatedElevator.ID] = updatedElevator.ElevatorStatus

		case reassignOrders := <-ch_motorTimedOut:
			for f := 0; f < config.N_FLOORS; f++ {
				for b := 0; b < config.N_BUTTONS-1; b++ {
					if reassignOrders.AssignedOrders[f][b] {
						ch_buttonPress <- elevio.ButtonEvent{f, elevio.ButtonType(b)}
					}
				}

			}
		}
	}
}

func AddElevator(ip string, elevator config.Elevator) {
	_, ok := allUpdatedElevators[ip]
	if ok == false {
		allUpdatedElevators[ip] = elevator
	}
}

// ------------------------------------------------------------------------------------------

//COST CALCULATION

func CalculateCost(buttonPress elevio.ButtonEvent) string {
	lowerCost := time.Duration(math.MaxInt64)
	lowerCostID := ""

	for k, e := range allUpdatedElevators {
		elev2 := config.Elevator{e.Floor, e.State, e.Direction, e.AssignedRequests, e.LightMatrix} //??
		elev2.AssignedRequests[buttonPress.Floor][buttonPress.Button] = true
		cost := TimeToIdle(elev2)
		if cost == lowerCost {
			lowerCostID = LowestNum(lowerCostID, k)
		} else if cost < lowerCost {
			lowerCost = cost
			lowerCostID = k
		}

	}
	return lowerCostID
}

func Requests_chooseDirection(elevator config.Elevator) elevio.MotorDirection {
	switch elevator.Direction {
	case elevio.MD_Up:
		if Fsm.IsOrderAbove(elevator) {
			return elevio.MD_Up
		} else if Fsm.IsOrderBelow(elevator) {
			return elevio.MD_Down
		}
		return elevio.MD_Stop
	case elevio.MD_Down:
		if Fsm.IsOrderBelow(elevator) {
			return elevio.MD_Down
		} else if Fsm.IsOrderAbove(elevator) {
			return elevio.MD_Up
		}
		return elevio.MD_Stop
	case elevio.MD_Stop:
		if Fsm.IsOrderAbove(elevator) {
			return elevio.MD_Up
		} else if Fsm.IsOrderBelow(elevator) {
			return elevio.MD_Down
		}
		return elevio.MD_Stop
	default:
		return elevio.MD_Stop
	}
}

func Requests_shouldStop(elevator config.Elevator) bool {
	switch elevator.Direction {
	case elevio.MD_Down:
		return elevator.AssignedRequests[elevator.Floor][elevio.BT_HallDown] || elevator.AssignedRequests[elevator.Floor][elevio.BT_Cab] || !Fsm.IsOrderBelow(elevator)
	case elevio.MD_Up:
		return elevator.AssignedRequests[elevator.Floor][elevio.BT_HallUp] || elevator.AssignedRequests[elevator.Floor][elevio.BT_Cab] || !Fsm.IsOrderAbove(elevator)
	case elevio.MD_Stop:
		return true
	default:
		return true
	}
}

func Request_clearAtCurrentFloor(e_old config.Elevator) config.Elevator {
	e := e_old
	for btn := 0; btn < config.N_BUTTONS; btn++ {
		if e.AssignedRequests[e.Floor][btn] {
			e.AssignedRequests[e.Floor][btn] = false
		}
	}
	return e
}

func TimeToIdle(elevator config.Elevator) time.Duration {
	duration := 0 * time.Millisecond

	switch elevator.State {
	case config.ES_IDLE:
		elevator.Direction = Requests_chooseDirection(elevator)
		if elevator.Direction == elevio.MD_Stop {
			return duration
		}
	case config.ES_MOVING:
		elevator.Floor += int(elevator.Direction)
		duration += config.TRAVEL_TIME / 2

	case config.ES_DOOROPEN:
		duration -= config.DOOR_OPEN_TIME / 2
		elevator.Direction = Requests_chooseDirection(elevator)
		if elevator.Direction == elevio.MD_Stop {
			return duration
		}
	}

	for {
		if Requests_shouldStop(elevator) {
			elevator = Request_clearAtCurrentFloor(elevator)
			duration += config.DOOR_OPEN_TIME
			elevator.Direction = Requests_chooseDirection(elevator)
			if elevator.Direction == elevio.MD_Stop {
				return duration
			}
		}
		elevator.Floor += int(elevator.Direction)
		duration += config.TRAVEL_TIME
	}
}

func LowestNum(num1 string, num2 string) string {
	n1, _ := strconv.ParseInt(num1, 10, 0)
	n2, _ := strconv.ParseInt(num2, 10, 0)
	if n1 < n2 {
		return num1
	} else {
		return num2
	}
}
