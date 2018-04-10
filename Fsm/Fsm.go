package Fsm

import(

	"../elevio"
	"fmt"
 	//"../config"
 	"time"
 	//"../Map"
 )

const (

	ES_INIT = 0 //temporary
	ES_IDLE = 1
	ES_MOVING = 2
	ES_DOOROPEN = 3
	ES_STUCK = 4 //stuck mellom to etasjer -- Hvordan vet vi den er stuck?
)

var ExecuteOrders [4][3] bool

var dir elevio.MotorDirection
var lastDirection elevio.MotorDirection //Kan hende denne er unødvendig
var state int
var lastFloor int



func Fsm(Ch_assignedOrders chan elevio.ButtonEvent, Ch_floor chan int, Ch_DoorTimeout chan bool) {
	Init()
	doortimer := time.NewTimer(3*time.Second)
	doortimer.Stop()

	for {
		switch(state){
		case ES_INIT:
			fmt.Println("State: Init")
		case ES_IDLE:
			fmt.Println("State: Idle")
		case ES_MOVING:
			fmt.Println("State: Moving")
		case ES_DOOROPEN:
			fmt.Println("State: DoorOpen")
		}

		switch(dir){
		case elevio.MD_Up:
			fmt.Println("Dir: Up")
		case elevio.MD_Down:
			fmt.Println("Dir: Down")
		case elevio.MD_Stop:
			fmt.Println("Dir: Stop")
		}

		select{
		case newOrder := <- Ch_assignedOrders:
			switch(state){	
				case ES_IDLE:
					if newOrder.Floor == lastFloor{
						ClearOrdersAtCurrentFloor(lastFloor)
						state = ES_DOOROPEN
						elevio.SetDoorOpenLamp(true)
						doortimer.Reset(3 * time.Second)
					}else{ 
						addOrder(newOrder)
						fmt.Println("Orders: %v", ExecuteOrders)
						executeOrder(lastFloor, newOrder.Floor)
						state = ES_MOVING
					}
				case ES_DOOROPEN:
					if newOrder.Floor == lastFloor{
						doortimer.Reset(3 * time.Second)
					}else{
						addOrder(newOrder)
					}

				default:
					addOrder(newOrder)
					fmt.Println("Orders: %v", ExecuteOrders) 
			}

		case reachedFloor := <- Ch_floor:
			elevio.SetFloorIndicator(reachedFloor)
			lastFloor = reachedFloor
			switch(state){
			case ES_INIT:
				elevio.SetMotorDirection(elevio.MD_Stop)
				lastDirection = dir
				dir = elevio.MD_Stop
				state = ES_IDLE

			case ES_MOVING:
				if (CheckOrdersAtFloor(reachedFloor)){
					//Kan lage funksjon av dette:
					lastDirection = dir
					dir = elevio.MD_Stop
					elevio.SetMotorDirection(dir)
					ClearOrdersAtCurrentFloor(lastFloor)
					elevio.SetDoorOpenLamp(true)
					doortimer.Reset(3 * time.Second)
					state = ES_DOOROPEN
				}else if (!CheckUpcomingFloors(reachedFloor)) {
					changeDirection()
					if (CheckOrdersAtFloor(reachedFloor)) {
						lastDirection = dir
						dir = elevio.MD_Stop
						elevio.SetMotorDirection(dir)
						ClearOrdersAtCurrentFloor(lastFloor)
						elevio.SetDoorOpenLamp(true)
						doortimer.Reset(3 * time.Second)
						state = ES_DOOROPEN
					}
				}

			default: 
				fmt.Println("ERROR. Reaching floor with unknown state")
			}

		case <-doortimer.C:
			fmt.Println("DoorTimeout")
			elevio.SetDoorOpenLamp(false)
			if CheckIfAnyOrders() {
				dir = lastDirection
				state = ES_MOVING
				//Koden under her kan forenkles: 
				if CheckUpcomingFloors(lastFloor){
					fmt.Println("inside doortimer - checkupcomingfloor")
					dir=lastDirection
					elevio.SetMotorDirection(dir)
				} else{
					changeDirection()
					fmt.Println("Inside doortimer - else statement")
					if (CheckUpcomingFloors(lastFloor)){

					dir=lastDirection
					fmt.Println("dir = ", dir)
					elevio.SetMotorDirection(dir)
					}
				}
			} else{
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
func addOrder(pressedButton elevio.ButtonEvent) bool{
	fmt.Println("Er inn i addOrder")
	if (ExecuteOrders[pressedButton.Floor][pressedButton.Button] == false){
		ExecuteOrders[pressedButton.Floor][pressedButton.Button] = true
		elevio.SetButtonLamp(pressedButton.Button, pressedButton.Floor, true)
		return true
	}
	return false

}

/*
func OnFloorArrival(newFloor int)  {
	lastFloor = newFloor

	if (state == ES_MOVING){
		if (ShouldStop(lastFloor)) {
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
			if (IsOrderAbove(floor)){
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
*/

func changeDirection(){
	if dir == elevio.MD_Up || lastFloor == 3{
		dir = elevio.MD_Down
	} else if dir == elevio.MD_Down || lastFloor == 0{
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
	for f := 0; f < 4; f++ {
		for b := 0; b < 3; b++ {
			if(ExecuteOrders[f][b] == true){
				fmt.Println("It has more orders to execute")
				return true
			}
		}
	}
	return false
}


func CheckOrdersAtFloor(floor int) bool {
	switch(dir){
	case elevio.MD_Up:
		return (ExecuteOrders[floor][elevio.BT_HallUp] || ExecuteOrders[floor][elevio.BT_Cab])

	case elevio.MD_Down: 
		return (ExecuteOrders[floor][elevio.BT_HallDown] || ExecuteOrders[floor][elevio.BT_Cab])

	case elevio.MD_Stop:
		return (ExecuteOrders[floor][elevio.BT_HallUp] || ExecuteOrders[floor][elevio.BT_HallDown] || ExecuteOrders[floor][elevio.BT_Cab])

	default:
		fmt.Println("ERROR: direction is stop")
		return false
	}
}

func CheckUpcomingFloors(lastFloor int) bool{
	switch(dir){
	case elevio.MD_Up:
		return IsOrderAbove(lastFloor)
	case elevio.MD_Down:
		return IsOrderBelow(lastFloor)
	}
	return false
}


func IsOrderAbove(lastFloor int) bool {
	if lastFloor == 3 {
		return false
	}
	for f := lastFloor + 1; f < 4; f++{
   		if CheckOrdersAtFloor(f){
   			fmt.Println("It has orders above")
   			return true
   		}
	}
	return false
}

func IsOrderBelow(lastFloor int) bool{
	if(lastFloor == 0){
		return false
	}
	for f := 0; f < lastFloor; f++{
		if CheckOrdersAtFloor(f){
			fmt.Println("It has orders below")
   			return true
   		}
	}
	return false
}

func ClearOrdersAtCurrentFloor(floor int) {
	fmt.Println("Er inne i ClearOrdersAtCurrentFloor")

	ExecuteOrders[floor][elevio.BT_Cab] = false
	elevio.SetButtonLamp(elevio.BT_Cab, floor, false)
	switch dir{
	case elevio.MD_Up:
		ExecuteOrders[floor][elevio.BT_HallUp] = false
		elevio.SetButtonLamp(elevio.BT_HallUp, floor, false)
		if (!IsOrderAbove(lastFloor)){
			ExecuteOrders[floor][elevio.BT_HallDown] = false
			elevio.SetButtonLamp(elevio.BT_HallDown, floor, false)
		}
	case elevio.MD_Down:
		ExecuteOrders[floor][elevio.BT_HallDown] = false
		elevio.SetButtonLamp(elevio.BT_HallDown, floor, false)
		if (!IsOrderBelow(lastFloor)){
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
