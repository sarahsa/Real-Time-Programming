package OrderManager

import (
	"fmt"
	"math"
	"../config"
	"../elevio"
	"../network/network/bcast"
	"../network/network/peers"
	"log"
	"io/ioutil"
	"os"
	"math/rand"
	"time"
	"strconv"
	"../Fsm"
	"../sync"
	"strings"
	"../backUp"
)

const (
	DOOR_OPEN_TIME time.Duration = 3000
	TRAVEL_TIME    time.Duration = 2500
	includeCab     bool          = false
)

//This map might be unnecessary, because its almost the same as the one above
var allUpdatedElevators = make(map[string]config.Elevator)

var TestElevatorID []string

var cabOrders[config.N_FLOORS]bool //save CabOrders

//var lights config.LightInfoPacket

var OrderRegistered [4][2]bool

func OrderManager(NewOrderTrans chan config.OrderPacket,
	NewOrderRecv chan config.OrderPacket,
	assignedOrders chan elevio.ButtonEvent,
	doorTimeout chan bool,
	ElevatorPacketTrans chan config.ElevatorStatusPacket,
	ElevatorPacketRecv chan config.ElevatorStatusPacket,
	ButtonPress chan elevio.ButtonEvent,
	myID string,
	hwPort string,
	UpdatedElevatorStatus chan config.Elevator,
	AckReceivedOrderRecv chan config.AcknowledgmentPacket,
	AckReceivedOrderTrans chan config.AcknowledgmentPacket,
	OrderIsExecuted chan elevio.ButtonEvent,
	AckExecutedRecv chan elevio.ButtonEvent,
	AckExecutedTrans chan elevio.ButtonEvent,
	MotorTimedOut chan config.OrderMatrix) {

	peerUpdateCh := make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool)
	//received := make(chan config.ReceivedAck)

	go peers.Transmitter(15847, myID, peerTxEnable) //15647 , 15670
	go peers.Receiver(15847, peerUpdateCh)          //15647

	go bcast.Transmitter(23232, NewOrderTrans, ElevatorPacketTrans, AckReceivedOrderTrans, AckExecutedTrans)
	go bcast.Receiver(23232, NewOrderRecv, ElevatorPacketRecv, AckReceivedOrderRecv, AckExecutedRecv)

	go elevio.PollButtons(ButtonPress)

	go sync.SendElevatorUpdate(allUpdatedElevators[myID],
		UpdatedElevatorStatus,
		ElevatorPacketTrans,
		myID)
	//go SendOrderUntilAck(ButtonPress, received)

	for {
		//sync.SyncAllLights(OrderRegistered)
		select {
		case p := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)

			if p.New != "" {
				TestElevatorID = append(TestElevatorID, p.New)
				addElevator(p.New, Fsm.GetElevatorStatus())
			}


			if len(p.Lost) ==1 && len(p.Peers) > 1 {
				for i := range p.Lost {
					fmt.Println("p.Lost= ", p.Lost[i])
					for f := 0; f < config.N_FLOORS; f++ {
						for b := 0; b < config.N_BUTTONS; b++ {
							if allUpdatedElevators[p.Lost[i]].AssignedRequests[f][b] == true{
								fmt.Println("CHECK")
								if elevio.ButtonType(b) != elevio.BT_Cab {
										ButtonPress <- elevio.ButtonEvent{f, elevio.ButtonType(b)}
								}
							}
						}
					}
				delete(allUpdatedElevators, p.Lost[i])
				}
			}

		case buttonPress := <-ButtonPress:
			fmt.Println("Button press at " + myID)
			fmt.Println(buttonPress)

			//Add Cab orders directly to Fsm
			if buttonPress.Button == elevio.BT_Cab {
				//Backup cabOrders to file
				cabOrders[buttonPress.Floor] = true
				backUp.SaveToDisk(buttonPress, cabOrders)
				assignedOrders <- buttonPress

			} else if len(allUpdatedElevators) > 1 {
				fmt.Println("CALCULATING COST")

				executer := CalculateCost(buttonPress)
				//executer := AssignOrderToRandomElevator()
				fmt.Println("Executer = ", executer)
				for _, e := range allUpdatedElevators{
					fmt.Println("AssignedRequests after cost: ", e.AssignedRequests)
				}
				NewOrderTrans <- config.OrderPacket{executer, buttonPress}
			}
			//received <- config.ReceivedAck{buttonPress, false}
		case recvButtonPacket := <-NewOrderRecv:
			AckReceivedOrderTrans <- config.AcknowledgmentPacket{myID, recvButtonPacket.Executer, recvButtonPacket.Button}


		case ack := <-AckReceivedOrderRecv:
			//received <- config.ReceivedAck{ack.Button, true}
			fmt.Println("Received ack from " + ack.Sender)

			//elevio.SetButtonLamp((ack.Button).Button, (ack.Button).Floor, true)


			fmt.Println("OrderRegistered when received: %v", OrderRegistered)
			if ack.Executer == myID {
				assignedOrders <- ack.Button
				fmt.Println("-------------- BUTTON -------------------")
				fmt.Println("Exceuter ", ack.Executer)
				fmt.Println("ButtenEvent", ack.Button.Floor, ack.Button)
				fmt.Println("-----------------------------------------")
			}

			if ack.Sender != myID {
				OrderRegistered[(ack.Button).Floor][int((ack.Button).Button)] = true
				sync.SyncAllLights(OrderRegistered)
				//elevio.SetButtonLamp((ack.Button).Button, (ack.Button).Floor, true)
			}

		case order := <-OrderIsExecuted:
			fmt.Println("ORDER IS EXECUTED")
			fmt.Println("Orderbutton: ", order.Button)
			fmt.Println("Setting Caborder to false") // MÅ FIKSES
			cabOrders[order.Floor] = false
			saveToDisk(order, cabOrders)

			if order.Button != elevio.BT_Cab {
				AckExecutedTrans <- order
			}


		case ackExecuted := <-AckExecutedRecv:
			//fmt.Print("ORDER IS ACKNOWLEDGMEN EXECUTED")
			fmt.Println("ackExecuted.Button: ", ackExecuted.Button)
			fmt.Println("OrderRegistered after execution: ", OrderRegistered)

			if OrderRegistered[ackExecuted.Floor][int(ackExecuted.Button)] == true {
				//elevio.SetButtonLamp(ackExecuted.Button, ackExecuted.Floor, false)
				OrderRegistered[ackExecuted.Floor][int(ackExecuted.Button)] = false
				sync.SyncAllLights(OrderRegistered)

			}
			//
			//

		case updatedElevator := <-ElevatorPacketRecv:
			allUpdatedElevators[updatedElevator.ID] = updatedElevator.ElevatorStatus

		case reassignOrders := <-MotorTimedOut:
			fmt.Println("MotorTimedOut channel")
			elem := allUpdatedElevators[myID]
			elem.State = Fsm.ES_STUCK
			allUpdatedElevators[myID] = elem
			for f := 0; f < config.N_FLOORS; f++ {
				for b := 0; b < config.N_BUTTONS-1; b++ {
					if reassignOrders.AssignedOrders[f][b] {
						fmt.Println("Reassigning")
						ButtonPress <- elevio.ButtonEvent{f, elevio.ButtonType(b)}
					}
				}

			}
		} // select
	} //for

} // ordermanagerfunc

func SendOrderUntilAck(ButtonPressed chan elevio.ButtonEvent, receivedAck chan config.ReceivedAck) {
	for {

		select {
		case recAck := <-receivedAck:
			//Do not send anymore
			if recAck.Status == false {
				ButtonPressed <- recAck.Button
			}
			//case <-time.After(time.Millisecond * 2000):

		}
	}

}

func IsElevatorNearest(myID string) bool {
	myCost := timeToIdle(allUpdatedElevators[myID])

	for k := range allUpdatedElevators {
		if k != myID {
			cost := timeToIdle(allUpdatedElevators[k])
			if cost < myCost {
				return false
			}
		}
	}

	return true
}

func FindNearestElevator(myID string) string {
	return "no"

}

func AssignOrderToRandomElevator() string {
	rand.Seed(time.Now().Unix())
	randomElev := TestElevatorID[rand.Intn(len(TestElevatorID))]
	fmt.Println("The elevator chosen is: %s", randomElev)
	return randomElev

}

func addElevator(ip string, elevator config.Elevator) {
	_, ok := allUpdatedElevators[ip]
	if ok == false {
		allUpdatedElevators[ip] = elevator
	}
}

func requests_chooseDirection(elevator config.Elevator) elevio.MotorDirection {
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
		}else{
			return elevio.MD_Stop}
	default:
		return elevio.MD_Stop
	}
}

func requests_shouldStop(elevator config.Elevator) bool {
	switch elevator.Direction {
	case elevio.MD_Down:
		return elevator.AssignedRequests[elevator.Floor][elevio.BT_HallDown] || elevator.AssignedRequests[elevator.Floor][elevio.BT_Cab] || !Fsm.IsOrderBelow(elevator)
	case elevio.MD_Up:
		return elevator.AssignedRequests[elevator.Floor][elevio.BT_HallUp] || elevator.AssignedRequests[elevator.Floor][elevio.BT_Cab] || !Fsm.IsOrderAbove(elevator)
	case elevio.MD_Stop:
		return true
		//return Fsm.CheckOrdersAtFloor(elevator.Floor)
	default:
		return true
	}
}

func request_clearAtCurrentFloor(e_old config.Elevator) config.Elevator {
	e := e_old
	for btn := 0; btn < config.N_BUTTONS; btn++ {
		if e.AssignedRequests[e.Floor][btn] {
			e.AssignedRequests[e.Floor][btn] = false
		}
	}
	return e
}

func timeToIdle(elevator config.Elevator) time.Duration {
	duration := 0 * time.Millisecond

	switch elevator.State {
	case Fsm.ES_IDLE:
		fmt.Println("Inside IDLE cost")
		elevator.Direction = requests_chooseDirection(elevator)
		if elevator.Direction == elevio.MD_Stop {
			return duration
		}
	case Fsm.ES_MOVING:
		fmt.Println("Inside Moving cost")
		elevator.Floor += int(elevator.Direction)
		duration += TRAVEL_TIME / 2

	case Fsm.ES_DOOROPEN:
		fmt.Println("Inside DOOROPEN cost")
		duration -= DOOR_OPEN_TIME / 2
		elevator.Direction = requests_chooseDirection(elevator)
		if elevator.Direction == elevio.MD_Stop {
			return duration
		}
	case Fsm.ES_STUCK:
		fmt.Println("Inside STUCK cost")
		duration = 2* DOOR_OPEN_TIME
		return duration

	}

	for {
		if requests_shouldStop(elevator) {
			elevator = request_clearAtCurrentFloor(elevator) //????
			duration += DOOR_OPEN_TIME
			elevator.Direction = requests_chooseDirection(elevator)
			if elevator.Direction == elevio.MD_Stop {
				return duration
			}
		}
		elevator.Floor += int(elevator.Direction)
		duration += TRAVEL_TIME
	}
}

func CalculateCost(buttonPress elevio.ButtonEvent) string {
	lowerCost := time.Duration(math.MaxInt64)
	lowerCostID := ""

	for k, e := range allUpdatedElevators {
		fmt.Println("Key: ",k,"e.state: ", e.State)
		if e.State != Fsm.ES_STUCK {
			if ((buttonPress.Button == elevio.BT_HallUp && e.AssignedRequests[buttonPress.Floor][elevio.BT_HallDown] != true) || ((buttonPress.Button == elevio.BT_HallDown && e.AssignedRequests[buttonPress.Floor][elevio.BT_HallUp] != true))){
				elev2 := config.Elevator{e.Floor, e.State, e.Direction, e.AssignedRequests, e.LightMatrix} //??
				elev2.AssignedRequests[buttonPress.Floor][buttonPress.Button] = true
				cost := timeToIdle(elev2)
				fmt.Println("-----------COST in FOR------------")
				fmt.Println("LowerCost: %d", cost)
				fmt.Println("LowerCostID: ", k)
				fmt.Println("-----------------------------------")

				if cost == lowerCost {
					lowerCostID = lowestNum(lowerCostID, k)
					fmt.Println("-----------COST------------")
					fmt.Println("LowerCost: %d", lowerCost)
					fmt.Println("LowerCostID: ", lowerCostID)
					fmt.Println("-----------------------")
				} else if cost < lowerCost {
					lowerCost = cost
					lowerCostID = k
					fmt.Println("-----------COST------------")
					fmt.Println("LowerCost: %d", lowerCost)
					fmt.Println("LowerCostID: ", lowerCostID)
					fmt.Println("-----------------------")
				}
			}
		}
	}
	return lowerCostID
}

func lowestNum(num1 string, num2 string) string {
	n1, _ := strconv.ParseInt(num1, 10, 0)
	n2, _ := strconv.ParseInt(num2, 10, 0)
	if n1 < n2 {
		return num1
	} else {
		return num2
	}
}

//-----------------------------------------------------------------------------------------------------

func saveToDisk(buttonPress elevio.ButtonEvent, cabOrders[config.N_FLOORS] bool){

	file, err := os.Create("backUp.txt")
	fmt.Println("backup created")
	if err != nil {
		log.Fatal("Cannot create file", err)
	}

	defer file.Close()

	for f := 0; f < config.N_FLOORS; f++ {
		order := strconv.FormatBool(cabOrders[f])
		_ , err = file.WriteString(order)
		_ , err = file.WriteString(" ")
			fmt.Println("writing to file")
			if err != nil {
				log.Fatal("Failed to backup", err)
			}
	}

	//save changes
	err = file.Sync()
	if err != nil{ log.Fatal("Cannot create file", err) }
}

func LoadFromDisk(e config.Elevator) {

	data, err := ioutil.ReadFile("backUp.txt")
	if err != nil {
		log.Fatal("Failed to read from backup", err)
	}
	//fmt.Println("fra load :", string(data))

	 backUpOrders := string(data) // srting
	 backUpOrdersList := strings.Split(backUpOrders, " ") // liste med string

	 for f := 0; f < config.N_FLOORS; f++ {
			 //order, _ := strconv.Atoi(string(buf[f]))
		 //fmt.Println(backUpOrders[f])
		 //fmt.Println("assigning orders from disk: ", backUpOrders)
		 fmt.Println("backUpOrdersList: ",backUpOrdersList[f])
		 order, _ := strconv.ParseBool(backUpOrdersList[f])
		 e.AssignedRequests[f][config.BT_CAB] = order
	 }

	fmt.Println("Loaded from disk array",e.AssignedRequests)
}

func intToBool(i int)bool  {
	if i == 1 {
		return true
	} else {
		return false
	}
}
