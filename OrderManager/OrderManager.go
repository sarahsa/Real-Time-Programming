package OrderManager

import (
      "../elevio"
       "../config"
      //"log"
      //"io/ioutil"
      //"os"
      //"time"
      //"../Fsm"
)

// key = string (ID/IP)
// value = Elevator
// get
var elevators = make(map[string]config.Elevator)

var ExecuteOrders[config.N_FLOORS][config.N_BUTTONS] bool

func executeOrder(buttonPress elevio.ButtonEvent, ) bool {


    cost := testCost()

    if (cost == true){
        AddOrder(buttonPress)
        return true
    }else{
      return false
    }
}

func addElevator(ip string, elevator config.Elevator)  {
  _, ok := elevators[ip]
  if (ok == false){
      elevators[ip] = elevator
  }
}



func testCost() bool {
  return true
}

//maa sammenligne kostene og legge til ordren dersom kosten returnerer true
func AddOrder(buttonPress elevio.ButtonEvent) bool {

    if (ExecuteOrders[buttonPress.Floor][buttonPress.Button] == false){
        ExecuteOrders[buttonPress.Floor][buttonPress.Button] = true
        return true
    }

    return false
}



/*/the cost for a single elevator
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
