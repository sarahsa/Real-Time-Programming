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

/*
const ClearRequestType {
    all = iota
    inDirn = iota,
}

ClearRequestType clearRequestType = ClearRequestType.inDirn;*/

var activeElevators = make(map[string]config.Elevator)

//This map might be unnecessary, because its almost the same as the one above
var allUpdatedElevators = make(map[string]config.Elevator)

/*
var ExecuteOrders[config.N_FLOORS][config.N_BUTTONS] bool  //TEST!!
var elev = make(chan config.Elevator)
*/
var TestElevatorID []string

func OrderManager(ButtonPacketTrans chan config.ButtonPressPacket,
	ButtonPacketRecv chan config.ButtonPressPacket,
	assignedOrders chan elevio.ButtonEvent,
	doorTimeout chan bool,
	ElevatorPacketTrans chan config.ElevatorStatusPacket,
	ElevatorPacketRecv chan config.ElevatorStatusPacket,
	ButtonPress chan elevio.ButtonEvent,
	myID string,
	hwPort string,
	UpdatedElevatorStatus chan config.Elevator) {

	peerUpdateCh := make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool)

	go peers.Transmitter(15847, myID, peerTxEnable) //15647 , 15670
	go peers.Receiver(15847, peerUpdateCh)          //15647

	go bcast.Transmitter(23232, ButtonPacketTrans, ElevatorPacketTrans)
	go bcast.Receiver(23232, ButtonPacketRecv, ElevatorPacketRecv)

	go elevio.PollButtons(ButtonPress)

	go sync.SendElevatorUpdate(activeElevators[myID],
		UpdatedElevatorStatus,
		ElevatorPacketTrans,
		myID)

	for {
		select {
		case p := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)

			if p.New != "" {
				TestElevatorID = append(TestElevatorID, p.New)
				addElevator(p.New, Fsm.GetElevatorStatus())
				fmt.Println("--------MapUpdate------")
				fmt.Println("Legger til ny heis")
				//elem := elevators[myID]
				fmt.Println("-----------------------")
			}

			if len(p.Lost) != 0 {
				for i := range p.Lost {
					delete(activeElevators, p.Lost[i])
				}
			}
			//Only for debugging purposes. Prints out the map, ie. elevator.
			for key, value := range activeElevators {
				fmt.Println("Key: ", key, "Value: ", value)
			}

		case buttonPress := <-ButtonPress:
			fmt.Println("Button press at " + myID)
			fmt.Println(buttonPress)

			//Add Cab orders directly to Fsm
			if buttonPress.Button == elevio.BT_Cab {
				//Backup cabOrders to file
				assignedOrders <- buttonPress

				//The order is a HallOrder, and must be assigned to an elevator.
			} else if len(activeElevators) > 1 {
				//assignedOrders <- buttonPress

				//This returns ID of the elevator which should execute the order.
				//executer := CalculateCost(buttonPress)
				executer := AssignOrderToRandomElevator()
				fmt.Println("Executer = ", executer)
				//elem :=allUpdatedElevators[executer]

				ButtonPacketTrans <- config.ButtonPressPacket{executer, buttonPress, true}

				//SyncInfo: Since the broadcasting happens in a separate thread. Use the
				//obtained info to assign order.

				//Calculate Cost
				//
			}
			//Receive the elevator information being broadcasted. It receives the updated
			//status on this channel. Should we add all the objects in a map? If so, when a peer
			//disappears it should also be removed from this map. Maybe we could just update the elevator map.
			//What about MotorTimeout??
		case updatedElevator := <-ElevatorPacketRecv:
			fmt.Println("updatedElevator_id: ", updatedElevator.ID)
			allUpdatedElevators[updatedElevator.ID] = updatedElevator.ElevatorStatus

			//debugging purposes
			for key, value := range allUpdatedElevators {
				fmt.Println("allUpdatedElevators: ")
				fmt.Println("Key: ", key, "Value: ", value)
			}

		case recvPacket := <-ButtonPacketRecv:
			//assignedOrders <- elevio.ButtonEvent{recvPacket.Floor, recvPacket.Button}
			fmt.Println("Received from " + recvPacket.Executer)
			//fmt.Println(recvPacket.Floor, " ", recvPacket.Button)

			if recvPacket.Executer == myID {
				assignedOrders <- recvPacket.Button
				fmt.Println("-------------- BUTTON -------------------")
				fmt.Println("Exceuter ", recvPacket.Executer)
				fmt.Println("ButtenEvent", recvPacket.Button.Floor, recvPacket.Button)
				fmt.Println("-----------------------------------------")
			} else if recvPacket.AssignedOrder == true { //HallOrderLamp on
				button := recvPacket.Button
				elevio.SetButtonLamp(button.Button, button.Floor, true)
			} /*else {
				button := recvPacket.Button
				if allUpdatedElevators[recvPacket.Executer].AssignedRequests[button.Floor][button.Button] == false {
					elevio.SetButtonLamp(button.Button, button.Floor, false)
					//bool = false
				}
			}*/

		}

	}

}

//debugging purposes
func printInfo(e config.Elevator) {
	fmt.Println("HEIS INFO")
	fmt.Println("Floor ", e.Floor)
	fmt.Println("State ", e.State)
	fmt.Println("Direction ", e.Direction)
	fmt.Println("Order ", e.AssignedRequests)
}

func IsElevatorNearest(myID string) bool {
	myCost := timeToIdle(activeElevators[myID])

	for k := range activeElevators {
		if k != myID {
			cost := timeToIdle(activeElevators[k])
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
	_, ok := activeElevators[ip]
	if ok == false {
		activeElevators[ip] = elevator
	}
}

func requests_chooseDirection(elevator config.Elevator) elevio.MotorDirection {
	switch elevator.Direction {
	case elevio.MD_Up:
		if Fsm.IsOrderAbove(elevator.Floor) {
			return elevio.MD_Up
		} else if Fsm.IsOrderBelow(elevator.Floor) {
			return elevio.MD_Down
		}
		return elevio.MD_Stop
	case elevio.MD_Down:
		if Fsm.IsOrderBelow(elevator.Floor) {
			return elevio.MD_Down
		} else if Fsm.IsOrderAbove(elevator.Floor) {
			return elevio.MD_Up
		}
		return elevio.MD_Stop
	case elevio.MD_Stop:
		if Fsm.IsOrderAbove(elevator.Floor) {
			return elevio.MD_Up
		} else if Fsm.IsOrderBelow(elevator.Floor) {
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
		return elevator.AssignedRequests[elevator.Floor][elevio.BT_HallDown] || elevator.AssignedRequests[elevator.Floor][elevio.BT_Cab] || !Fsm.IsOrderBelow(elevator.Floor)
	case elevio.MD_Up:
		return elevator.AssignedRequests[elevator.Floor][elevio.BT_HallUp] || elevator.AssignedRequests[elevator.Floor][elevio.BT_Cab] || !Fsm.IsOrderAbove(elevator.Floor)
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
		elev2 := config.Elevator{e.Floor, e.State, e.Direction, e.AssignedRequests} //??
		elev2.AssignedRequests[buttonPress.Floor][buttonPress.Button] = true
		fmt.Println("ELEV 2:", elev2)
		cost := timeToIdle(elev2)
		fmt.Println("-----------COST in FOR------------")
		fmt.Println("LowerCost: %d", cost)
		fmt.Println("LowerCostID: ", k)
		fmt.Println("-----------------------")
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
