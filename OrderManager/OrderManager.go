package OrderManager

import (
	"fmt"
	"math"

	"../config"
	"../elevio"
	//"flag"
	"../network/network/bcast"
	"../network/network/peers"
	// "../network/network/localip"
	//"log"
	//"io/ioutil"
	//"os"
	"math/rand"
	"time"

	"strconv"

	"../Fsm"
	"../sync"
)

const (
	DOOR_OPEN_TIME time.Duration = 3000
	TRAVEL_TIME    time.Duration = 2500
	includeCab     bool          = false
)

//This map might be unnecessary, because its almost the same as the one above
var allUpdatedElevators = make(map[string]config.Elevator)

var activeElevators = make(map[string]config.Elevator)

var TestElevatorID []string

var lights config.LightInfoPacket

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

	go peers.Transmitter(15847, myID, peerTxEnable) //15647 , 15670
	go peers.Receiver(15847, peerUpdateCh)          //15647

	go bcast.Transmitter(23232, NewOrderTrans, ElevatorPacketTrans, AckReceivedOrderTrans, AckExecutedTrans)
	go bcast.Receiver(23232, NewOrderRecv, ElevatorPacketRecv, AckReceivedOrderRecv, AckExecutedRecv)

	go elevio.PollButtons(ButtonPress)

	go sync.SendElevatorUpdate(allUpdatedElevators[myID],
		UpdatedElevatorStatus,
		ElevatorPacketTrans,
		myID)

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
				activeElevators[p.New] = allUpdatedElevators[p.New]
			}

			if len(p.Lost) != 0 {
				for i := range p.Lost {
					delete(allUpdatedElevators, p.Lost[i])
				}
			}
			/*
				//Only for debugging purposes. Prints out the map, ie. elevator.
				for key, value := range allUpdatedElevators {
					fmt.Println("Key: ", key, "Value: ", value)
				}*/

		case buttonPress := <-ButtonPress:
			fmt.Println("Button press at " + myID)
			fmt.Println(buttonPress)

			//Add Cab orders directly to Fsm
			if buttonPress.Button == elevio.BT_Cab {
				//Backup cabOrders to file
				assignedOrders <- buttonPress

			} else if len(activeElevators) > 1 {

				executer := CalculateCost(buttonPress)
				//executer := AssignOrderToRandomElevator()
				fmt.Println("Executer = ", executer)
				NewOrderTrans <- config.OrderPacket{executer, buttonPress}

			}
		case recvButtonPacket := <-NewOrderRecv:
			fmt.Println("REceived new order")
			AckReceivedOrderTrans <- config.AcknowledgmentPacket{myID, recvButtonPacket.Executer, recvButtonPacket.Button}

		case ack := <-AckReceivedOrderRecv:
			fmt.Println("Received ack from " + ack.Sender)
			if ack.Sender != myID {
				OrderRegistered[(ack.Button).Floor][int((ack.Button).Button)] = true
				elevio.SetButtonLamp((ack.Button).Button, (ack.Button).Floor, true)
			}

			fmt.Println("OrderRegistered: %v", OrderRegistered)
			if ack.Executer == myID {
				assignedOrders <- ack.Button
				fmt.Println("-------------- BUTTON -------------------")
				fmt.Println("Exceuter ", ack.Executer)
				fmt.Println("ButtenEvent", ack.Button.Floor, ack.Button)
				fmt.Println("-----------------------------------------")
			}

		case order := <-OrderIsExecuted:
			fmt.Print("ORDER IS EXECUTED")
			AckExecutedTrans <- order

		case ackExecuted := <-AckExecutedRecv:
			fmt.Print("ORDER IS ACKNOWLEDGMEN EXECUTED")
			fmt.Println("ackExecuted.Button: ", ackExecuted.Button)
			elevio.SetButtonLamp(ackExecuted.Button, ackExecuted.Floor, false)
			OrderRegistered[ackExecuted.Floor][int(ackExecuted.Button)] = false

		case updatedElevator := <-ElevatorPacketRecv:
			allUpdatedElevators[updatedElevator.ID] = updatedElevator.ElevatorStatus

		case reassignOrders := <-MotorTimedOut:

			for f := 0; f < config.N_FLOORS; f++ {
				for b := 0; b < config.N_BUTTONS-1; b++ {
					if reassignOrders.AssignedOrders[f][b] {
						ButtonPress <- elevio.ButtonEvent{f, elevio.ButtonType(b)}
					}
				}

			}
		}

	}
}

/*func printInfo(e config.Elevator) {
	fmt.Println("HEIS INFO")
	fmt.Println("Floor ", e.Floor)
	fmt.Println("State ", e.State)
	fmt.Println("Direction ", e.Direction)
	fmt.Println("Order ", e.AssignedRequests)
}*/

/*func ManageOldElevator(key string, pos elevio.ButtonEvent, value bool) {
	for k, v := range oldElevators {
		if k == key {
			v.AssignedOrders[pos.Floor][int(pos.Button)] = value
		}
	}
}

func ClearOrder() {
	light := false
	for f := 0; f < config.N_FLOORS; f++ {
		for b := 0; b < config.N_BUTTONS-1; b++ {
			if LightMatrix[f][b] {
				for k, v := range allUpdatedElevators {
					if v.AssignedRequests[f][b] {
						light = true
					}
					button := elevio.ButtonEvent{f, elevio.ButtonType(b)}
					ManageOldElevator(k, button, false)
				}
				if light == false {
					elevio.SetButtonLamp(elevio.ButtonType(b), f, false)
					LightMatrix[f][b] = false
				} else {
					light = false
				}
			}
		}
	}
}*/

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

/*func addLightMatrix(ip string, matrix[4][3]bool]) {
	_, ok := activeElevators[ip]
	if ok == false {
		activeElevators[ip] = matrix
	}
}*/

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
		}
		return elevio.MD_Stop
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
			/*if (onClearedRequest){
			    onClearedRequest(b)
			}*/
		}
	}
	return e
}

/*func request_clearAtCurrentFloor(elev chan config.Elevator) Elevator {
    switch elev.config.onClearedRequestVariant {
    case CV_ALL:
        for btn Button := 0; btn < N_BUTTONS; btn++  {
            e.AssignedRequests[e.Floor][btn] = 0
        }
        break
    case CV_InDirn:
        e.AssignedRequests[e.Floor][BT_CAB] = 0
        switch e.dirn {
        case D_Up:
            e.AssignedRequests[e.Floor][B_HallUp] = 0
            if !Fsm.IsOrderAbove(e) {
                e.AssignedRequests[e.Floor][B_HallDown] = 0
            }
            break
        case D_Down:
            e.AssignedRequests[e.Floor][B_HallDown] = 0
            if (!Fsm.IsOrderBelow(e)) {
                e.AssignedRequests[e.Floor][B_HallUp] = 0
            }
            break
        case D_Stop:
        configault:
            e.AssignedRequests[e.Floor][B_HallUp] = 0
            e.AssignedRequests[e.Floor][B_HallDown] = 0
            break
        }
        break
    configault:
        break
    }
    return e
}*/

func timeToIdle(elevator config.Elevator) time.Duration {
	duration := 0 * time.Millisecond

	switch elevator.State {
	case Fsm.ES_IDLE:
		elevator.Direction = requests_chooseDirection(elevator)
		if elevator.Direction == elevio.MD_Stop {
			return duration
		}
	case Fsm.ES_MOVING:
		elevator.Floor += int(elevator.Direction)
		duration += TRAVEL_TIME / 2

	case Fsm.ES_DOOROPEN:
		duration -= DOOR_OPEN_TIME / 2
		elevator.Direction = requests_chooseDirection(elevator)
		if elevator.Direction == elevio.MD_Stop {
			return duration
		}

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
		fmt.Println("ELEV 1", e)
		elev2 := config.Elevator{e.Floor, e.State, e.Direction, e.AssignedRequests, e.LightMatrix} //??
		elev2.AssignedRequests[buttonPress.Floor][buttonPress.Button] = true
		fmt.Println("ELEV 2:", elev2)
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

		fmt.Println("AFTER COST ELEV 1", e)
		//fmt.Println(elev2)
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

/*func executeOrder(buttonPress elevio.ButtonEvent, ) bool {

    cost := testCost()
    if (cost == true){
        AddOrder(buttonPress)
        return true
    }else{
      return false
    }
}*/

//maa sammenligne kostene og legge til ordren dersom kosten returnerer true
/*func AddOrder(buttonPress elevio.ButtonEvent) bool {

    if (ExecuteOrders[buttonPress.Floor][buttonPress.Button] == false){
        ExecuteOrders[buttonPress.Floor][buttonPress.Button] = true
        return true
    }
    return false
}*/

/*func testCost() bool {
  return true
}*/

/*func saveToDisk(filname string) error{

	data, err := jason.Marshal( []byte, error )
		if err != nil{
			log.Println("Eroor: Failed to backup")
			return err
		}

        //func WriteFile(filename string, data []byte, perm os.FileMode) error
		if err := ioutil.WriteFile(filename, data , 0644); err != nil { // writes to file and checks for returned error
			log.Println("Error: Failed to backup")
			return err
		}
		return nil
}*/

/*func loadFromDisk(filename string) error { //func Stat(name string) (FileInfo, error)

    var queue

    if _, err := os.State(filename); err == nil {
        data, err := ioutil.ReadFile(filename)
        if err != nil{
        log.Println("Error: Failed to read from backup")
            return err
    }

    if err := jason.Unmarshal(data,queue); err != nil {
        log.Println("Error: failed to Unmarshal")
    }

    }

    return nil

}*/

//-----------------------------------------------------------------------------------------------------

// COST

//the cost for a single elevator (MORTENFYHN)
/*func cost(buttonPress elevio.ButtonEvent, e Elevator) int {

    cost := 0
    floor := e.floor
    dir :=  e.dir
    targetfloor := buttonPress.Floor

    if floor == -1 { //between floors
        cost++
    }

    else if Fsm.State != elevio.MD_Stop {
        cost += 2
    }

    floor, dir = incrementFloor(floor, dir)

    //simulates the elevator cost until it reaches the target floorm max 10 simulations
    for n := 0; !(floor == targetfloor && CheckOrdersAtFloor(floor)) && n < 10; n++ {
        if  Fsm.CheckOrdersAtFloor(floor){
            cost += 2
            elevio.SetButtonLamp(BT_HallDown, floor, false)
            elevio.SetButtonLamp(BT_HallUp, floor, false)
            elevio.SetButtonLamp(BT_Cab, false)
            elevio.SetDoorOpenLamp(true)
        }
        dir = chooseDirection(floor, dir)
        floor, dir = incrementFloor(floor, dir)
        cost += 2
    }

    return cost
} */
/*func incrementFloor(floor int, dir int) (int, int) {

    switch dir {
    case elevio.MD_Down:
            floor--
        case elevio.MD_Up:
            floor++
    }

    if floor <= 0 && dir == elevoi.MD_Down{
        dir = elevio.MD_Up
        floor = 0
    }

    if floor >= NumFloors - 1 && dir == elevoi.MD_Up{
        dir = elevio.MD_Down
        floor = NumFloors - 1
    }
    return floor, dir
}*/

/*func request_clearAtCurrentFloor(elev chan config.Elevator) Elevator {
    switch elev.config.onClearedRequestVariant {
    case CV_ALL:
        for btn Button := 0; btn < N_BUTTONS; btn++  {
            e.AssignedRequests[e.Floor][btn] = 0
        }
        break
    case CV_InDirn:
        e.AssignedRequests[e.Floor][BT_CAB] = 0
        switch e.dirn {
        case D_Up:
            e.AssignedRequests[e.Floor][B_HallUp] = 0
            if !Fsm.IsOrderAbove(e) {
                e.AssignedRequests[e.Floor][B_HallDown] = 0
            }
            break
        case D_Down:
            e.AssignedRequests[e.Floor][B_HallDown] = 0
            if (!Fsm.IsOrderBelow(e)) {
                e.AssignedRequests[e.Floor][B_HallUp] = 0
            }
            break
        case D_Stop:
        configault:
            e.AssignedRequests[e.Floor][B_HallUp] = 0
            e.AssignedRequests[e.Floor][B_HallDown] = 0
            break
        }
        break
    configault:
        break
    }
    return e
}*/

//NEED TO BE FIXED, void delegate(CallType c) onClearedRequest = null
