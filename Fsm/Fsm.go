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
/*
*****Mulig struct, i stedet for globale variabler*****
type Elevator struct{
	assignedOrders[4][3] bool
	dir elevio.MotorDirection
	state int
	floor int
}
*/

var ExecuteOrders [4][3] bool

var dir elevio.MotorDirection
var lastDirection elevio.MotorDirection //Kan hende denne er unødvendig
var state int
var lastFloor int
floors := make(chan int)


func Fsm(Ch_assignedOrders chan elevio.ButtonEvent, Ch_DoorTimeout chan bool) {
	Init()
	go elevio.PollFloorSensor(floors)
	doortimer := time.NewTimer(3*time.Second)
	doortimer.Stop()

	for {
		//Only for debugging purposes
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

		switch(lastDirection){
		case elevio.MD_Up:
			fmt.Println("LastDir: Up")
		case elevio.MD_Down:
			fmt.Println("LastDir: Down")
		case elevio.MD_Stop:
			fmt.Println("LastDir: Stop")
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
				if reachedFloor == 3 || reachedFloor == 0{
					fmt.Println("Changing direction due to 0 or 3")
					changeDirection()
				}
				if (CheckOrdersAtFloor(reachedFloor)){
					//Kan lage funksjon av dette:
					lastDirection = dir
					dir = elevio.MD_Stop
					elevio.SetMotorDirection(dir)
					ClearOrdersAtCurrentFloor(lastFloor)
					elevio.SetDoorOpenLamp(true)
					doortimer.Reset(3 * time.Second)
					state = ES_DOOROPEN
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
					fmt.Println("dir = ", dir)
					elevio.SetMotorDirection(dir)
					}else{
						dir = lastDirection
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
	fmt.Println("doorclosed")
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

func changeDirection(){
	fmt.Println("Changing direction ")
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
/*
func UpdateElevator(elevator chan config.Elevator)  {
	Ch_floor := make(chan int)
	go elevio.PollFloorSensor(Ch_floor)
	elevator.floor = <- Ch_floor
	elevator.state = state
	elevator.direction = dir
	elevator.requests = ExecuteOrders
}*/
