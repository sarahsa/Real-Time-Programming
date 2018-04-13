package OrderManager

import (
      "../elevio"
      "../config"
      "fmt"
      //"flag"
      "../network/network/peers"
      "../network/network/bcast"
     // "../network/network/localip"
      //"log"
      //"io/ioutil"
      //"os"
      //"time"
      "../Fsm"
)

const (

DOOR_OPEN_TIME int = 3000
TRAVEL_TIME int = 2500
includeCab bool = false

)

const ClearRequestType {
    all = iota
    inDirn = iota,
}

ClearRequestType clearRequestType = ClearRequestType.inDirn;

var elevators = make(map[string]config.Elevator)

var ExecuteOrders[config.N_FLOORS][config.N_BUTTONS] bool  //TEST!!

var elev = make(chan config.Elevator)

func OrderManager(ButtonPacketTrans chan config.ButtonPressPacket, ButtonPacketRecv chan config.ButtonPressPacket, assignedOrders chan elevio.ButtonEvent, doorTimeout chan bool, ElevatorTrans chan config.Elevator, ElevatorRecv chan config.Elevator, ButtonPress chan elevio.ButtonEvent, myID string, hwPort string)  {

    peerUpdateCh := make(chan peers.PeerUpdate)
    peerTxEnable := make(chan bool)

    go peers.Transmitter(15647, myID, peerTxEnable) //15647 , 15670
    go peers.Receiver(15647, peerUpdateCh) //15647

    go bcast.Transmitter(23232, ButtonPacketTrans, ElevatorTrans)
    go bcast.Receiver(23232, ButtonPacketRecv, ElevatorRecv)

    go elevio.PollButtons(ButtonPress)


    for{
    		select{
    		case p := <-peerUpdateCh:
    			fmt.Printf("Peer update:\n")
    			fmt.Printf("  Peers:    %q\n", p.Peers)
    			fmt.Printf("  New:      %q\n", p.New)
    			fmt.Printf("  Lost:     %q\n", p.Lost)

                if(p.New != "" && myID == p.New){
                    addElevator(myID, Fsm.GetElevatorStatus())
                    fmt.Println("Legger til ny heis")
                    elem := elevators[myID]
                    fmt.Println("maps : ", elem.Floor)
                }

    		case buttonPress := <-ButtonPress:
    			fmt.Println("Button press at " + myID)
    			fmt.Println(buttonPress)
                if buttonPress.Button == elevio.BT_CAB  {   //Differentiating between cab and hall orderss
                    assignedOrders <- buttonPress
                    fmt.Println("CabOrder added")

                }else{
                    ButtonPacketTrans <- config.ButtonPressPacket{myID, buttonPress.Floor, buttonPress.Button}

                }

    		case recvPacket := <-ButtonPacketRecv:
                //assignedOrders <- elevio.ButtonEvent{recvPacket.Floor, recvPacket.Button}

        		fmt.Println("Received from " + recvPacket.Sender)
        		fmt.Println(recvPacket.Floor, " ", recvPacket.Button)


                /*orderAccepted, changeMade := OrderManager.AddOrder(buttonPress)
                fmt.Println("Change Made: ", changeMade)
                if changeMade {
                    ButtonPacketRecv <- ButtonPressPacket{myID, buttonPress.Floor, int(buttonPress.Button)}
                }

            case recvElevPacket := <- ElevatorRecv:

                for k, v := range OrderManager.elevators {
                    if k != recvPacket.ID {
                        elevators[recvPacket.ID] = recvPacket
                    }
                }*/
    	}

    }

}

func addElevator(ip string, elevator config.Elevator)  {
  _, ok := elevators[ip]
  if (ok == false){
      elevators[ip] = elevator
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

func requests_chooseDirection(elev chan config.Elevator) MotorDirection {
    switch elev.Direction {
    case elevio.MD_Up:
        if Fsm.IsOrderAbove(elev.Floor) {
            return elevio.MD_Up
        } else if Fsm.IsOrderBelow(elev.Floor) {
            return elevio.MD_Down
        }
        return elevio.MD_Stop
    case elevio.MD_Down:
        if Fsm.IsOrderBelow(elev.Floor) {
            return elevio.MD_Down
        } else if Fsm.IsOrderAbove(elev.Floor) {
            return elevio.MD_Up
        }
        return elevio.MD_Stop
    case elevio.MD_Stop:
        if Fsm.IsOrderAbove(elev.Floor) {
            return elevio.MD_Up
        } else if Fsm.IsOrderBelow(elev.Floor){
            return elevio.MD_Down:
        }
        return elevio.MD_Stop
    configault:
        return elevio.MD_Stop
    }
}

func requests_shouldStop(elev chan config.Elevator) bool {
    switch elev.Direction {
    case elevio.MD_Down:
        return elev.AssignedRequests[elev.Floor][elevio.BT_HallDown] || elev.AssignedRequests[elev.Floor][elevio.BT_CAB] || !Fsm.IsOrderBelow(elev.Floor)
    case elevio.MD_Up:
        return elev.AssignedRequests[elev.Floor][elevio.BT_HallUp] || elev.AssignedRequests[elev.Floor][elevio.BT_CAB] || !Fsm.IsOrderAbove(elev.Floor)
    case elevio.MD_Stop:
    configault:
        return true
    }
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

//NEED TO BE FIXED, void delegate(CallType c) onClearedRequest = null
func request_clearAtCurrentFloor(e_old Elevator, onClearedRequest(b Button, floor int)) Elevator {
    e Elevator := e_old
    for btn Button := 0; btn < N_BUTTONS; btn++ {
        if (e.AssignedRequests[e.Floor][btn]){
            e.AssignedRequests[e.floor][btn] = 0;
            if (onClearedRequest){
                onClearedRequest(btn, floor)
            }
        }
    }
    return e
}

func timeToIdle(elev chan config.Elevator) int{
    duration int = 0

    switch elev.State {
    case ES_IDLE:
        elev.Direction = requests_chooseDirection(elev)
        if elev.Direction == elevio.MD_Stop {
            return duration
        }
        break
    case ES_MOVING:
        duration += TRAVEL_TIME/2
        elev.Floor += elev.Direction
        break
    case ES_DOOROPEN:
        duration -= DOOR_OPEN_TIME/2
    }

    for {
        if requests_shouldStop(elev) {
            elev = request_clearAtCurrentFloor(elev, nil) //????
            duration += DOOR_OPEN_TIME
            elev.Direction = requests_chooseDirection(elev)
            if elev.Direction == elevio.MD_Stop {
                return duration
            }
        }
        elev.Floor += elev.Direction
        duration += TRAVEL_TIME
    }
}
