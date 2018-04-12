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

// key = string (ID/IP)
// value = Elevator
// get
var elevators = make(map[string]config.Elevator)

var ExecuteOrders[config.N_FLOORS][config.N_BUTTONS] bool  //TEST!!

var elev = make(chan config.Elevator)

func OrderManager(ButtonPacketTrans chan config.ButtonPressPacket, ButtonPacketRecv chan config.ButtonPressPacket,
     assignedOrders chan elevio.ButtonEvent, doorTimeout chan bool, ElevatorTrans chan config.Elevator,
     ElevatorRecv chan config.Elevator, ButtonPress chan elevio.ButtonEvent, myID string, hwPort string)  {

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

    			ButtonPacketTrans <- config.ButtonPressPacket{myID, buttonPress.Floor, buttonPress.Button}

    		case recvPacket := <-ButtonPacketRecv:
                assignedOrders <- elevio.ButtonEvent{recvPacket.Floor, recvPacket.Button}

        		fmt.Println("Received from " + recvPacket.Sender)
        		fmt.Println(recvPacket.Floor, " ", recvPacket.Button)

    	}

    }

    /*for{
            select{
            case p := <-peerUpdateCh:
                fmt.Printf("Peer update:\n")
                fmt.Printf("  Peers:    %q\n", p.Peers)
                fmt.Printf("  New:      %q\n", p.New)
                fmt.Printf("  Lost:     %q\n", p.Lost)

            case buttonPress := <-ButtonPress:
                fmt.Println("Button press at " + myID)
                fmt.Println(buttonPress)		myID = fmt.Sprintf("peer-%s", localIP)

                //floor <- Ch_floor
                //OrderManager.executeOrder(buttonPress)
                //fmt.Println("Change Made: ", changeMade)

                ButtonPacketTrans <- config.ButtonPressPacket{myID, buttonPress.Floor, buttonPress.Button}
                //ElevatorTrans <- Ch_elevator

            case recvPacket := <-ButtonPacketRecv:
                assignedOrders <- elevio.ButtonEvent{recvPacket.Floor, recvPacket.Button}

                fmt.Println("Received from " + recvPacket.Sender)
                fmt.Println(recvPacket.Floor, " ", recvPacket.Button)

            	orderAccepted, changeMade := OrderManager.AddOrder(buttonPress)
                fmt.Println("Change Made: ", changeMade)
                if changeMade {
                    ButtonPacketRecv <- ButtonPressPacket{myID, buttonPress.Floor, int(buttonPress.Button)}
                }

            case recvElevPacket := <- ElevatorRecv:

                for k, v := range OrderManager.elevators {
                    if k != recvPacket.ID {
                        elevators[recvPacket.ID] = recvPacket
                    }
                }

            }

    }*/

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
}

//maa sammenligne kostene og legge til ordren dersom kosten returnerer true
func AddOrder(buttonPress elevio.ButtonEvent) bool {

    if (ExecuteOrders[buttonPress.Floor][buttonPress.Button] == false){
        ExecuteOrders[buttonPress.Floor][buttonPress.Button] = true
        return true
    }

    return false
}*/

//-----------------------------------------------------------------------------------------------------

/*
func testCost() bool {
  return true
}
//the cost for a single elevator
func cost(buttonPress elevio.ButtonEvent, e Elevator) int {

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
}



func requests_chooseDirection(e Elevator) MotorDirection {
    switch e.dir {
    case elevio.MD_Up:
        if IsOrderAbove(e.floor) {
            return elevio.MD_Up
        } else if IsOrderBelow(e.floor) {
            return elevio.MD_Down
        }
        return elevio.MD_Stop
    case elevio.MD_Down:
        if IsOrderBelow(e.floor) {
            return elevio.MD_Down
        } else if IsOrderAbove(e.floor) {
            return elevio.MD_Up
        }
        return elevio.MD_Stop
    case elevio.MD_Stop:
        if IsOrderAbove(e.floor) {
            return elevio.MD_Up
        } else if IsOrderBelow(e.floor){
            return elevio.MD_Down:
        }
        return elevio.MD_Stop
    configault:
        return elevio.MD_Stop
    }
}

func incrementFloor(floor int, dir int) (int, int) {

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
}

/*
func OrderManager()  {

}


func saveToDisk(filname string) error{

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
}

func loadFromDisk(filename string) error { //func Stat(name string) (FileInfo, error)

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

}

//-------------------------------------------------------------------------------------

func requests_shouldStop(e Elevator) bool {
    switch e.dir {
    case MD_Down:
        return e.requests[e.floor][BT_HallUp] || e.requests[e.floor][BT_Cab] || !IsOrderBelow(e.floor)
    case MD_Up:
        return e.requests[e.floor][BT_HallUp] || e.requests[e.floor][BT_Cab] || !IsOrderAbove(e.floor)
    case MD_Stop:
    configault:
        return true
    }
}

func request_clearAtCurrentFloor(e Elevator) Elevator {
    switch e.config.onClearedRequestVariant {
    case CV_ALL:
        for btn Button := 0; btn < N_BUTTONS; btn++  {
            e.requests[e.floor][btn] = 0
        }
        break
    case CV_InDirn:
        e.requests[e.floor][B_Cab] = 0
        switch e.dirn {
        case D_Up:
            e.requests[e.floor][B_HallUp] = 0
            if !requests_above(e) {
                e.requests[e.floor][B_HallDown] = 0
            }
            break
        case D_Down:
            e.requests[e.floor][B_HallDown] = 0
            if (!requests_below(e)) {
                e.requests[e.floor][B_HallUp] = 0
            }
            break
        case D_Stop:
        configault:
            e.requests[e.floor][B_HallUp] = 0
            e.requests[e.floor][B_HallDown] = 0
            break
        }
        break
    configault:
        break
    }
    return e
}

func request_clearAtCurrentFloor(e_old Elevator, onClearedRequest(b Button, floor int)) Elevator {
    e Elevator := e_old
    for btn Button := 0; btn < N_BUTTONS; btn++ {
        if (e.requests[e.floor][btn]){
            e.requests[e.floor][btn] = 0;
            if (onClearedRequest){
                onClearedRequest(btn, floor)
            }
        }
    }
    return e
}

func timeToIdle(e Elevator) int{
    duration int = 0

    switch e.behaviour {
    case EB_Idle:
        e.dirn = requests_chooseDirection(e)
        if e.dirn == D_Stop {
            return duration
        }
        break
    case EB_Moving:
        duration += TRAVEL_TIME/2
        e.floor += e.dirn
        break
    case EB_DoorOpen:
        duration -= DOOR_OPEN_TIME/2
    }

    for {
        if requests_shouldStop(e) {
            e = request_clearAtCurrentFloor(e, nil)
            duration += DOOR_OPEN_TIME
            e.dirn = requests_chooseDirection(e)
            if e.dirn == D_Stop {
                return duration
            }
        }
        e.floor += e.direction
        duration += TRAVEL_TIME
    }
}
*/
