package Fsm

import(

	"../elevio"
	"fmt"
 	"../config"
 	"time"
 	"../Map"
 )

const (

	ES_INIT = 0 //temporary
	ES_IDLE = 1
	ES_MOVING = 2
	ES_DOOROPEN = 3
	ES_STUCK = 4 //stuck mellom to etasjer -- Hvordan vet vi den er stuck?
)

//Denne er her kun for aa teste
floors = 4
buttons = 3
var ExecuteOrders [floors][buttons] bool

var dir elevio.MotorDirection
var lastDirection int //Kan hende denne er unødvendig
var state int
var lastFloor int



func Fsm(Ch_assignedOrders chan elevio.buttonEvent, Ch_floor chan int, Ch_DoorTimeout chan bool) {
	Init()

	for {
		select{
		case newOrder := <- Ch_assignedOrders:
			switch(state){	
				case ES_IDLE: 
					addOrder(newOrder)
					executeOrder(lastFloor, newOrder.Floor)
					state = ES_MOVING
				case ES_DOOROPEN:
					//reset timer
			
				default:
					addOrder(newOrder) 
			}

		case reachedFloor := <- Ch_floor:
			lastFloor = reachedFloor
			switch(state){
			case ES_INIT:
				elevio.SetMotorDirection(elevio.MD_Stop)
				lastDirection = dir
				dir = elevio.MD_Stop
				state = ES_IDLE

			case ES_MOVING:
				//Sjekke om bestillinger i nådd etasje
				if (CheckOrdersAtFloor(lastFloor)){
					//Kan lage funksjon av dette:
					lastDirection = dir
					dir = elevio.MD_Stop
					elevio.SetMotorDirection(dir)
					//door open
					elevio.SetDoorOpenLamp(true)
					//start doorTimer
					state = ES_DOOROPEN
					ClearOrdersAtCurrentFloor(lastFloor)
				}


			default: 
				//Noe gikk galt. Print ERROR
			}

		case <- Ch_DoorTimeout:
			elevio.SetDoorOpenLamp(false)
			if (CheckIfAnyOrders()) {
				dir = lastDirection
				if (CheckOrdersAtFloor(lastFloor)) {
					//Kan lage funksjon av dette:
					lastDirection = dir
					dir = elevio.MD_Stop
					elevio.SetMotorDirection(dir)
					//door open
					elevio.SetDoorOpenLamp(true)
					//start timer
					state = ES_DOOROPEN
					ClearOrdersAtCurrentFloor(lastFloor)
				}

				else if (CheckUpcomingFloors(lastFloor)){
					dir=lastDirection
					elevio.SetMotorDirection(dir)
				}
				else{
					changeDirection()
					if (CheckUpcomingFloors(lastFloor)){
					dir=lastDirection
					elevio.SetMotorDirection(dir)
					}
				}

			}
			else{
				state = ES_IDLE
			}

		}
	}
}

func Init(){

	fmt.Println("Initializing...")
	elevio.SetDoorOpenLamp(false)
	state = ES_INIT

	for f := 0; f < 4; f++ {
		for b := elevio.ButtonType(0); b < 3; b++ {
    		elevio.SetButtonLamp(b, f, false)
		}
	}


	if elevio.GetFloor() == -1{
		elevio.SetMotorDirection(elevio.MD_Up)
		dir = elevio.MD_Up
	}

}

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
			dir = elevio.MotorDirection.MD_Stop
			//stopper heisen?
			elevio.SetMotorDirection(dir)


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

	switch dir {
	case elevio.MD_Up:
		return (floor == (floors - 1)) || ExecuteOrders[floor][elevio.BT_HallUp] || ExecuteOrders[floor][elevio.BT_Cab] || !IsOrdersAbove(floor)
	case elevio.MD_Down:
		return (floor == 0) || ExecuteOrders[floor][elevio.BT_Cab] || ExecuteOrders[floor][elevio.BT_HallDown] || !IsOrdersUnder(floor)
	case elevio.MD_Stop:
		return CheckOrdersAtFloor()
	}

	return false
}

//returnerer retning
func ChooseDirection(floor int) int {
	fmt.Println("Er inn i ChooseDirection")

	switch dir{
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

func changeDirection(){
	if dir == elevio.MD_Up {
		dir = elevio.MD_Down
	}
	else if dir == elevio.MD_Down{
		dir = elevio.MD_Up
	}
}

func executeOrder(lastFloor int, targetFloor int){

	if(lastFloor < targetFloor){
		dir = elevio.MD_Up

	} else if (lastFloor > targetFloor) {
		dir = elevio.MD_Down
		
	} else {								//lastFloor == targetFloor
		lastDirection = dir
		dir = elevio.MD_Stop
		//Fjerne ordre
		//Slå av lys
		ClearOrdersAtCurrentFloor(lastFloor)
		//Åpne dør
		elevio.SetDoorOpenLamp(true)
		//Start Timer
		dir = lastDirection
		
	}
	elevio.SetMotorDirection(dir)
}

func CheckIfAnyOrders() bool{
	for r := 0; r < 3; r++ {
		for c := 0; c < 4; c++ {
			if(ExecuteOrders[r][c] == true){
				return true
			}
		}
	}
	return false
}


func CheckOrdersAtFloor(floor int) bool {
	fmt.Println("Er inn i CheckOrdersAtFloor")
	switch(dir){
	case elevio.MD_Up:
		return (ExecuteOrders[floor][elevio.BT_HallUp] || ExecuteOrders[floor][elevio.BT_Cab])

	case elevio.MD_Down: 
		return (ExecuteOrders[floor][elevio.BT_HallDown] || ExecuteOrders[floor][elevio.BT_Cab])

	case elevio.MD_Stop:
		return (ExecuteOrders[floor][elevio.BT_HallUp] || ExecuteOrders[floor][elevio.BT_HallDown] || ExecuteOrders[floor][elevio.BT_Cab])

	default:
		fmt.Println("ERROR: direction is stop")

	}
}

func CheckUpcomingFloors(lastFloor int) bool{
	switch(dir){
	case elevio.MD_Up:
		return IsOrderAbove(lastFloor)
	case elevio.MD_Down:
		return IsOrderBelow(lastFloor)
	}
}


func IsOrderAbove(lastFloor int) bool {
	fmt.Println("Inni isOrdersAbove")
	if(lastFloor == 3){
		return false
	}
	for f := currentFloor + 1; f < floors; f++{
   		return CheckOrdersAtFloor(f))
   }
}

func IsOrderBelow(lastFloor int) bool{
	fmt.Println("Inn i isOrdersUnder")
	if(lastFloor == 0){
		return false
	}
	for f := 0; f < currentFloor; f++{
		return CheckOrdersAtFloor(f)
	}
}

func ClearOrdersAtCurrentFloor(floor int) {
	fmt.Println("Er inne i ClearOrdersAtCurrentFloor")

	ExecuteOrders[floor][elevio.BT_Cab] = false
	elevio.SetButtonLamp(elevio.BT_Cab, floor, false)
	switch dir{
	case elevio.MD_Up:
		ExecuteOrders[floor][elevio.BT_HallUp] = false
		elevio.SetButtonLamp(elevio.BT_HallUp, floor, false)
		if (!IsOrdersAbove){
			ExecuteOrders[floor][elevio.BT_HallDown] = false
			elevio.SetButtonLamp(elevio.BT_HallDown, floor, false)
		}
	case elevio.MD_Down:
		ExecuteOrders[floor][elevio.BT_HallDown] = false
		elevio.SetButtonLamp(elevio.BT_HallDown, floor, false)
		if (!IsOrdersUnder){
			ExecuteOrders[floor][elevio.BT_HallDown] = false
			elevio.SetButtonLamp(elevio.BT_HallDown, floor, false)
		}
	default:
		ExecuteOrders[floor][elevio.BT_HallUp] = false
		elevio.SetButtonLamp(elevio.BT_HallUp, floor, false)
		ExecuteOrders[floor][elevio.BT_HallDown] = false
		elevio.SetButtonLamp(elevio.BT_HallDown, floor, false)
	}
}

func DoorTimeout()  {
	fmt.Println("Er inne i DoorTimeout")

	if (state == ES_DOOROPEN){
		//direction = ChooseDirection(currentFloor)
		elevio.SetDoorOpenLamp(false)
		elevio.SetMotorDirection(dir)

		if (dir == elevio.MD_Stop){
			state = ES_IDLE
		} else {
			state = ES_MOVING
		}
	}
}
