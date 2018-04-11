
package map

import "encoding/json"
import "../elevio"
import "log"
import "io/ioutil"
import "os"
import "time"

const NumFloors = 4
const NumButtonsTypes = 3

// ta imot bestilling
// starte en timer når bestillinge er mottatt
// oppdatere backup med ny bestilling
// posisjonene/ bestillinger til alle heiser er sync
// regne ut cost funksjonen med ny bestilling (hvilken heis har lavest cost)
// deligerer den nye bestillinger til en annen heis hvis den har lavere cost
// sync alle bestillinger
// kaller på FSM og utfører bestilling

// -----------------------------------
// Feilhåndtering:
// - Timer går ut = bestilling har ikke blitt utført
// 		 Må sende bestillingen til heisen som skal utøre den på nytt

// -----------------------------------
// Trenger:
// - Gamle og nye bestillinger
// - Poisisjonene til alle aktive heiser
// - Hver heis har sin egen cost-variabel

// -----------------------------------
// Forskjell på nye bestillinger-array og skal utføre-array?

//save backup to disk

var dir* = Fsm.dir

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

type Elevator struct {

    MY_ID string
    Position int
    Direction int

}

//the cost for a single elevator
func cost(Ch_Buttons chan elevio.ButtonEvent, packet chan Elevator) int {

    cost := 0
    floor := Elevator.Position
    dir :=  Elevator.Direction
    targetfloor := elevio.ButtonEvent.Floor

    if floor == -1 { //between floors
        cost++
    }

    else if fsm.State != MD_Stop {
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
        }
        //dir = chooseDirection(floor, dir)
        floor, dir = incrementFloor(floor, dir)
        cost += 2
    }

    return cont

}

func incrementFloor(floor int, dir int) (int, int) {

    switch dir {
        case fsm.Motordirection.MD_Down:
            floor--
        case fsm.Motordirection.MD_Up:
            floor++
    }

    if floor <= 0 && dir == fsm.Motordirection.MD_Down{
        dir = fsm.Motordirection.MD_Up
        floor = 0
    }

    if floor >= NumFloors - 1 && dir == fsm.Motordirection.MD_Up{
        dir = fsm.Motordirection.MD_Down
        floor = NumFloors - 1
    }
    return floor, dir
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
// ------------------------------------

type Elevator struct{
    floor int
    dirn dirn
    requests[NumFloors][NumButtonsTypes] int
    behaviour ElevatorBehavior
}

type queue struct{
    matrix [NumFloors][NumButtonsTypes]bool
}

//Matrix that contains all local orders from all elevators
var local queue

//Matrix that contains orders that har been assigned to spesific elevators
//after orders have gone throw the cost function
var execute queue

//gets the input parameters from the CH-    Orders-channele, or something similare
func AddLocalOrder(floor int, button int){
    local[floor][button] = 1;
}

func AddExecuteOrder(floor int, button int,) {
    alreadyExist := IsExecuteOrder(floor, button)
    if !alreadyExist{
        exectue.setOrder(floor,button)
    }
}

func setOrder(floor int, button int){
    matrix[floor][button] = 1
}

//if elevator dead, reassign all orders to the network
func ReassignOrders(floor int, button int) {}

//--------------------------------------------------------

func RemoveExecuteOrders(floor int){
    for b := 0; b < NumButtonsTypes; b++ {
        exectue.setOrder(floor, b, 0)
    }
}

func RemoveOrders(floor int) {
    for b :=; b < NumButtonsTypes; b++ {
        local.setOrder(floor, b, 0)
        exectue.setOrder(floor, b, 0)
    }
}
