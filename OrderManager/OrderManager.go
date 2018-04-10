package OrderManager

import (
      "../elevio"
      //"log"
      //"io/ioutil"
      //"os"
      //"time"
      //"../Fsm"
)

const NumFloors = 4
const NumButtonsTypes = 3
var ExecuteOrders[4][3] bool


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

func AddOrder(buttonPress elevio.ButtonEvent)  (bool, bool) {

    //floor := <- bt.Floor

    if (ExecuteOrders[buttonPress.Floor][buttonPress.Button] == false){
        ExecuteOrders[buttonPress.Floor][buttonPress.Button] = true
        return true, true
    }

    return false, false
}


//lytter fra chan i main og oppdaterer etasje
func FloorUpdate(floor chan int)  {

}

//lytter fra chan i main og oppdaterer retning
func DirectionUpdate(direction chan elevio.MotorDirection){

}

//??
func LampUpdate()  {

}

func IsElevatorAlive()  bool {

}

//-------------------------------------------------------------------------------------

type Elevator struct {
    floor int
    dirn Dirn
    request[N_FLOORS][N_BUTTONS]
    behaviour ElevatorBehavior
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
