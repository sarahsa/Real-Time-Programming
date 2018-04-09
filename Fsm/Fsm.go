package Fsm

import(

	"../elevio"
	"fmt"
 	"../config"
 	"time"
 	"../Map2"
 )

const (

	//ES_Init = 0 //temporary
	ES_IDLE = 1
	ES_MOVING = 2
	ES_DOOROPEN = 3
	ES_STUCK = 4 //stuck mellom to etasjer -- Hvordan vet vi den er stuck?
)

/*
type Heartbeat struct {
	Position int //Floor
	State int
	ID string  //IP
	Direction elevio.MotorDirection
}*/

//Denne er her kun for aa teste
floors = 4
buttons = 3
var ExecuteOrders [floors][buttons] bool

var direction int
var state int
var currentFloor int



func Fsm(Ch_buttonEvent chan elevio.buttonEvent, Ch_floor chan int) {


	direction = elevio.MD_Stop
	state = ES_IDLE

	for {
		select{
		case event := <- Ch_buttonEvent:
			if (event.Button == elevio.BT_Cab){

			}

		case f := <- Ch_Floors:
			fmt.Println("Inni Ch_Floors")
			elevio.SetFloorIndicator(f)
			lastFloor = f

			switch state{

			case ES_Init:

				elevio.SetMotorDirection(elevio.MD_Stop)
				state = ES_Idle

			case ES_Moving:

				if CheckOrdersAtFloor(f){

					elevio.SetMotorDirection(elevio.MD_Stop)
					elevio.SetButtonLamp(elevio.BT_HallUp, f, false)

				}

				if (checkOrderAtGivenDirection(lastDirection,f)){
					elevio.SetMotorDirection(lastDirection)
				}

		}
	}
}
}

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

/*//riktig?
func OnInitBetweenFloors()  {
	fmt.Println("Er inn i OnInitBetweenFloors")
	elevio.SetMotorDirection(MD_Down)
	direction = elevio.MD_Down
	state = ES_MOVING
}*/

//returner true dersom ordren er lagt til, false ellers
func addOrder(pressedButton elevio.buttonEvent) bool{
	fmt.Println("Er inn i addOrder")
	if (ExecuteOrders[pressedButton.Floor][pressedButton.ButtonType] == false){
		ExecuteOrders[pressedButton.Floor][pressedButton.ButtonType] = true
		return true
	}
	return false

}


func OnFloorArrival(newFloor int)  {
	lastFloor = newFloor

	if (state == ES_MOVING){
		if (ShouldStop(floor)) {
			direction = elevio.MotorDirection.MD_Stop
			//stopper heisen?
			elevio.SetMotorDirection(direction)


			elevio.SetDoorOpenLamp(true)


			//sletter ordren
			ClearOrdersAtCurrentFloor(currentFloor)

			state = ES_DOOROPEN
		}
	}
}

//finner ut on heisen har grunn til stoppe/burde den stoppe
func ShouldStop(floor int) bool{
	fmt.Println("Er inn i ShouldStop")

	switch direction {
	case elevio.MD_Up:
		return (floor == (floors - 1)) || ExecuteOrders[floor][elevio.BT_HallUp] || ExecuteOrders[floor][elevio.BT_Cab] || !IsOrdersAbove(floor)
	case elevio.MD_Down:
		return (floor == 0) || ExecuteOrders[floor][elevio.BT_Cab] || ExecuteOrders[floor][elevio.BT_HallDown] || !IsOrdersUnder(floor)
	case elevio.MD_Stop:
		return CheckOrdersAtFloor()
	}

	return false
}

/*//Bra for en heis
func ShouldStop(floor int) bool {
	fmt.Println("Er inn i ShouldStop")
	if (CheckOrdersAtFloor(floor)){
		return true
	}else if (ChooseDirection(floor) == elevio.MotorDirection.MD_Stop){
		return true
	}else if (floor == 3 || floor == 0){
		return true
	}
	return false

}*/


//returnerer retning
func ChooseDirection(floor int) int {
	fmt.Println("Er inn i ChooseDirection")

	switch direction{
	case elevio.MD_Up:
			if (IsOrdersAbove(floor)){
				return elevio.MD_Up
			}else if (IsOrdersUnder(floor)){
				return elevio.MD_Down
			}else{
				return elevio.MD_Stop
			}

		case elevio.MD_Down:
			if (IsOrdersUnder(floor)){
				return elevio.MD_Down
			}else if (IsOrdersAbove(floor)){
				return elevio.MD_Up
			}else{
				return elevio.MD_Stop
			}

		case elevio.MD_Stop:
			if (IsOrdersAbove(floor)) {
				return elevio.MD_Up
			} else if (IsOrdersUnder(floor)){
				return elevio.MD_Down
			}
	}
	return elevio.MD_Stop
}


func CheckOrdersAtFloor(floor int) bool {
	fmt.Println("Er inn i CheckOrdersAtFloor")
	switch
	return (ExecuteOrders[floor][elevio.BT_HallUp] || ExecuteOrders[floor][elevio.BT_HallDown] || ExecuteOrders[floor][elevio.BT_Cab])
}


func IsOrdersAbove(currentFloor int) bool {
	fmt.Println("Inni isOrdersAbove")
   for f := currentFloor + 1; f < floors; f++{
   		return CheckOrdersAtFloor(f))
   }
}

func IsOrdersUnder(currentFloor int) bool{
	fmt.Println("Inn i isOrdersUnder")
	for f := 0; f < currentFloor; f++{
		return CheckOrdersAtFloor(f)
	}
}

func ClearOrdersAtCurrentFloor(floor int) {
	fmt.Println("Er inne i ClearOrdersAtCurrentFloor")

	ExecuteOrders[floor][elevio.BT_Cab] = 0
	switch direction{
	case elevio.MD_Up:
		ExecuteOrders[floor][elevio.BT_HallUp] = 0
		if (!IsOrdersAbove){
			ExecuteOrders[floor][elevio.BT_HallDown] = 0
		}
	case elevio.MD_Down:
		ExecuteOrders[floor][elevio.BT_HallDown] = 0
		if (!IsOrdersUnder){
			ExecuteOrders[floor][elevio.BT_HallDown] = 0
		}
	default:
		ExecuteOrders[floor][elevio.BT_HallUp] = 0
		ExecuteOrders[floor][elevio.BT_HallDown] = 0
	}
}

func DoorTimeout()  {
	fmt.Println("Er inne i DoorTimeout")

	if (state == ES_DOOROPEN){
		//direction = ChooseDirection(currentFloor)
		elevio.SetDoorOpenLamp(false)
		elevio.SetMotorDirection(direction)

		if (direction == elevio.MD_Stop){
			state = ES_IDLE
		} else {
			state = ES_MOVING
		}
	}
}
