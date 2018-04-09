package Fsm

import(

	"../elevio"
	"fmt"
 	"../config"
 	"time"
 	"../Map2"
 )

const (

	ES_Init = 0 //temporary
	ES_IDLE = 1
	ES_MOVING = 2
	ES_DOOROPEN = 3
	ES_STUCK = 4 //stuck mellom to etasjer
)

type Heartbeat struct {
	Position int //Floor
	State int
	ID string  //IP
	Direction elevio.MotorDirection
}

//Denne er her kun for aa teste
var ExecuteOrders [4][3] bool


var direction int
var state int = ES_Init
var floor int 


//trenger vi denne?
func Init(){

	fmt.Println("Initializing...")
	elevio.SetDoorOpenLamp(false)
	state = ES_Init

	for f := 0; f < 4; f++ {
		for b := elevio.ButtonType(0); b < 3; b++ {
    		elevio.SetButtonLamp(b, f, false)
		}
	}


	if elevio.GetFloor() == -1{
		elevio.SetMotorDirection(elevio.MD_Up)
		direction = elevio.MD_Up
	}

}



func Fsm(Ch_buttonEvent chan elevio.buttonEvent) {
	direction = elevio.MotorDirection.MD_Stop
	state = ES_IDLE

	for {
		select{
		case msg := <- Ch_buttonEvent:


		}
	}


}



//riktig? Maa fikse paa den
func OnInitBetweenFloors()  {
	fmt.Println("Er inn i OnInitBetweenFloors")
	elevio.SetMotorDirection(MD_Down)
	direction = elevio.MD_Down
	state = ES_MOVING

}

//Den burde vaare i Map, denne er kun bra for 1 heis
func addOrder(pressedButton elevio.buttonEvent) bool{
	fmt.Println("Er inn i addOrder")
	if (ExecuteOrders[pressedButton.Floor][pressedButton.ButtonType] == false){
		ExecuteOrders[pressedButton.Floor][pressedButton.ButtonType] = true
		return true
	}
	return false

}


func OnFloorArrival(newFloor int)  {
	fmt.Println("Er inn i OnFloorArrival")
	floor = newFloor

	elevio.SetFloorIndicator(floor)

	if (state = ES_MOVING){
		if (ShouldStop(floor)) {
			direction = elevio.MotorDirection.MD_Stop
			//fjerne ordre
			//dor aapenlampen skal vaare pa
			//Opner dora + timer
		}
	}

}


//Bra for en heis, men ikke for flere
func ShouldStop(floor int) bool {
	fmt.Println("Er inn i ShouldStop")

	switch dir{
	case elevio.MotorDirection.MD_Up:
		return !isOrdersUnder || ExecuteOrders[floor][elevio.ButtonType.BT_HallUp] || ExecuteOrders[floor][elevio.ButtonType.BT_Cab]
	case elevio.MotorDirection.MD_Down:
		return !isOrdersUnder || ExecuteOrders[floor][elevio.ButtonType.BT_HallDown] || ExecuteOrders[floor][elevio.ButtonType.BT_Cab]
	}

}

func ChooseDirection(floor int) int {

	fmt.Println("Er inn i ChooseDirection")

	switch dir{
		case elevio.MotorDirection.MD_Up:

			if (IsOrdersAbove(floor)){
				dir = elevio.MotorDirection.MD_Up
			}else if (IsOrdersUnder(floor)){
				dir = elevio.MotorDirection.MD_Down
			}else{
				dir = elevio.MotorDirection.MD_Stop
			}

		case elevio.MotorDirection.MD_Down:

			if (IsOrdersUnder(floor)){
				dir = elevio.MotorDirection.MD_Down
			}else if(IsOrdersAbove(floor)){
				dir = elevio.MotorDirection.MD_Up
			}else{
				dir = elevio.MotorDirection.MD_Stop
			}
		default:
			dir = elevio.MotorDirection.MD_Stop


	}
}


//sjekker for en heis!!
func CheckOrdersAtFloor(floor int) bool {
	fmt.Println("Er inn i CheckOrdersAtFloor")
	return (ExecuteOrders[f][0] || ExecuteOrders[f][1] || ExecuteOrders[f][2])
}

// Maa gjore det mer generelt
func IsOrdersAbove(currentFloor int) bool {
	fmt.Println("Inni isOrdersAbove")
   for f := currentFloor + 1; f < 4; f++{
   		if (CheckOrdersAtFloor(f)){
   			return true
   		}
   		return false
   }
}

func IsOrdersUnder(currentFloor int) bool{
	fmt.Println("Inn i isOrdersUnder")
	for f := 0; f < currentFloor; f++{
		if CheckOrdersAtFloor(f){
			return true
		}
		return false
		
	}
}
