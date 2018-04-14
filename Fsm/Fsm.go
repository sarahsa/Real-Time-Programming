package Fsm

import(

	"../elevio"
	"fmt"
 	"../config"
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

var elevator config.Elevator
var lastDirection elevio.MotorDirection //Kan hende denne er unødvendig
var floors = make(chan int)


func Fsm(Ch_assignedOrders chan elevio.ButtonEvent, Ch_DoorTimeout chan bool, Ch_UpdateElevatorStatus chan config.Elevator) {
	Init()
	go elevio.PollFloorSensor(floors)
	doortimer := time.NewTimer(3*time.Second)
	doortimer.Stop()

	for {
		//Only for debugging purposes
		switch(elevator.State){
		case ES_INIT:
			fmt.Println("elevator.elevator.State: Init")
		case ES_IDLE:
			fmt.Println("State: Idle")
		case ES_MOVING:
			fmt.Println("State: Moving")
		case ES_DOOROPEN:
			fmt.Println("State: DoorOpen")
		}

		switch(elevator.Direction){
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
			switch(elevator.State){
				case ES_IDLE:
					if newOrder.Floor == elevator.Floor{
						ClearOrdersAtCurrentFloor(elevator.Floor)
						elevator.State = ES_DOOROPEN
						elevio.SetDoorOpenLamp(true)
						doortimer.Reset(3 * time.Second)

					}else{
						addOrder(newOrder)
						fmt.Println("Orders: %v", elevator.AssignedRequests)
						executeOrder(elevator.Floor, newOrder.Floor)
						elevator.State = ES_MOVING
					}
				case ES_DOOROPEN:
					if newOrder.Floor == elevator.Floor{
						doortimer.Reset(3 * time.Second)
					}else{
						addOrder(newOrder)
					}

				default:
					addOrder(newOrder)
					fmt.Println("Orders: %v", elevator.AssignedRequests)
			}

		case reachedFloor := <- floors:
			elevio.SetFloorIndicator(reachedFloor)
			elevator.Floor = reachedFloor
			//fmt.Println("floor from reachedFloor: ", elevator.Floor)
			switch(elevator.State){
			case ES_INIT:
				elevio.SetMotorDirection(elevio.MD_Stop)
				lastDirection = elevator.Direction
				elevator.Direction = elevio.MD_Stop
				elevator.State = ES_IDLE
				Ch_UpdateElevatorStatus <- elevator

			case ES_MOVING:
				if reachedFloor == 3 || reachedFloor == 0{
					fmt.Println("Changing direction due to 0 or 3")
					changeDirection()
				}

				if (CheckOrdersAtFloor(reachedFloor)){
					//Kan lage funksjon av dette:
					lastDirection = elevator.Direction
					elevator.Direction = elevio.MD_Stop
					elevio.SetMotorDirection(elevator.Direction)
					ClearOrdersAtCurrentFloor(elevator.Floor)
					elevio.SetDoorOpenLamp(true)
					doortimer.Reset(3 * time.Second)
					elevator.State = ES_DOOROPEN
					Ch_UpdateElevatorStatus <- elevator
				}else{
					changeDirection()
					if CheckOrdersAtFloor(reachedFloor) {
						lastDirection = elevator.Direction
						elevator.Direction = elevio.MD_Stop
						elevio.SetMotorDirection(elevator.Direction)
						ClearOrdersAtCurrentFloor(elevator.Floor)
						elevio.SetDoorOpenLamp(true)
						doortimer.Reset(3 * time.Second)
						elevator.State = ES_DOOROPEN
						Ch_UpdateElevatorStatus <- elevator
					}
				}


			default:
				fmt.Println("ERROR. Reaching floor with unknown state")
			}
			//Ch_UpdateElevatorStatus <- elevator

		case <-doortimer.C:
			fmt.Println("DoorTimeout")
			elevio.SetDoorOpenLamp(false)
			if CheckIfAnyOrders() {
				elevator.Direction = lastDirection
				elevator.State = ES_MOVING
				//Koden under her kan forenkles:
				if CheckUpcomingFloors(elevator.Floor){
					fmt.Println("inside doortimer - checkupcomingfloor")
					elevator.Direction=lastDirection
					elevio.SetMotorDirection(elevator.Direction)
				} else{
					changeDirection()
					fmt.Println("Inside doortimer - else statement")
					if (CheckUpcomingFloors(elevator.Floor)){
					fmt.Println("dir = ", elevator.Direction)
					elevio.SetMotorDirection(elevator.Direction)
					}else{
						elevator.Direction = lastDirection
						elevio.SetMotorDirection(elevator.Direction)
					}
				}
			} else{
				elevator.State = ES_IDLE
			}

		}
	}
}

func Init(){

	fmt.Println("Initializing...")
	elevio.SetDoorOpenLamp(false)
	fmt.Println("doorclosed")
	elevator.State = ES_INIT

	for f := 0; f < 4; f++ {
		for b := elevio.ButtonType(0); b < 3; b++ {
    		elevio.SetButtonLamp(b, f, false)
		}
	}


	if elevio.GetFloor() == -1{
		elevio.SetMotorDirection(elevio.MD_Up)
		elevator.Direction = elevio.MD_Up
	}else{
		elevator.Floor = elevio.GetFloor()
	}


}

//returner true dersom ordren er lagt til, false ellers
func addOrder(pressedButton elevio.ButtonEvent) bool{
	fmt.Println("Er inn i addOrder")
	if (elevator.AssignedRequests[pressedButton.Floor][pressedButton.Button] == false){
		elevator.AssignedRequests[pressedButton.Floor][pressedButton.Button] = true
		elevio.SetButtonLamp(pressedButton.Button, pressedButton.Floor, true)
		return true
	}
	return false
}

func changeDirection(){
	fmt.Println("Changing direction ")
	if elevator.Direction == elevio.MD_Up || elevator.Floor == 3{
		elevator.Direction = elevio.MD_Down
	} else if elevator.Direction == elevio.MD_Down || elevator.Floor == 0{
		elevator.Direction = elevio.MD_Up
	}
}

func executeOrder(floor int, targetFloor int){

	if(floor < targetFloor){
		elevator.Direction = elevio.MD_Up

	} else if (floor > targetFloor) {
		elevator.Direction = elevio.MD_Down

	} else {								//elevator.Floor == targetFloor
		lastDirection = elevator.Direction
		elevator.Direction = elevio.MD_Stop
		//Fjerne ordre
		//Slå av lys
		ClearOrdersAtCurrentFloor(floor)
		//Åpne dør
		elevio.SetDoorOpenLamp(true)
		//Start Timer
		elevator.Direction = lastDirection

	}
	elevio.SetMotorDirection(elevator.Direction)
}

func CheckIfAnyOrders() bool{
	for f := 0; f < 4; f++ {
		for b := 0; b < 3; b++ {
			if(elevator.AssignedRequests[f][b] == true){
				fmt.Println("It has more orders to execute")
				return true
			}
		}
	}
	return false
}


func CheckOrdersAtFloor(floor int) bool {
	switch(elevator.Direction){
	case elevio.MD_Up:
		return (elevator.AssignedRequests[floor][elevio.BT_HallUp] || elevator.AssignedRequests[floor][elevio.BT_Cab])

	case elevio.MD_Down:
		return (elevator.AssignedRequests[floor][elevio.BT_HallDown] || elevator.AssignedRequests[floor][elevio.BT_Cab])

	case elevio.MD_Stop:
		return (elevator.AssignedRequests[floor][elevio.BT_HallUp] || elevator.AssignedRequests[floor][elevio.BT_HallDown] || elevator.AssignedRequests[floor][elevio.BT_Cab])

	default:
		fmt.Println("ERROR: direction is stop")
		return false
	}
}

func CheckUpcomingFloors(floor int) bool{
	switch(elevator.Direction){
	case elevio.MD_Up:
		return IsOrderAbove(floor)
	case elevio.MD_Down:
		return IsOrderBelow(floor)
	}
	return false
}


func IsOrderAbove(floor int) bool {
	if floor == 3 {
		return false
	}
	for f := floor + 1; f < 4; f++{
   		if CheckOrdersAtFloor(f){
   			fmt.Println("It has orders above")
   			return true
   		}
	}
	return false
}

func IsOrderBelow(floor int) bool{
	if(floor == 0){
		return false
	}
	for f := 0; f < floor; f++{
		if CheckOrdersAtFloor(f){
			fmt.Println("It has orders below")
   			return true
   		}
	}
	return false
}

func ClearOrdersAtCurrentFloor(floor int) {
	//fmt.Println("Er inne i ClearOrdersAtCurrentFloor")

	elevator.AssignedRequests[floor][elevio.BT_Cab] = false
	elevio.SetButtonLamp(elevio.BT_Cab, floor, false)
	switch elevator.Direction{
	case elevio.MD_Up:
		elevator.AssignedRequests[floor][elevio.BT_HallUp] = false
		elevio.SetButtonLamp(elevio.BT_HallUp, floor, false)
		if (!IsOrderAbove(elevator.Floor)){
			elevator.AssignedRequests[floor][elevio.BT_HallDown] = false
			elevio.SetButtonLamp(elevio.BT_HallDown, floor, false)
		}
	case elevio.MD_Down:
		elevator.AssignedRequests[floor][elevio.BT_HallDown] = false
		elevio.SetButtonLamp(elevio.BT_HallDown, floor, false)
		if (!IsOrderBelow(elevator.Floor)){
			elevator.AssignedRequests[floor][elevio.BT_HallDown] = false
			elevio.SetButtonLamp(elevio.BT_HallDown, floor, false)
		}
	default:
		elevator.AssignedRequests[floor][elevio.BT_HallUp] = false
		elevio.SetButtonLamp(elevio.BT_HallUp, floor, false)
		elevator.AssignedRequests[floor][elevio.BT_HallDown] = false
		elevio.SetButtonLamp(elevio.BT_HallDown, floor, false)
	}
}


func DoorTimeout()  {
	fmt.Println("Er inne i DoorTimeout")

	if (elevator.State == ES_DOOROPEN){
		//direction = ChooseDirection(currentFloor)
		elevio.SetDoorOpenLamp(false)
		elevio.SetMotorDirection(elevator.Direction)

		if (elevator.Direction == elevio.MD_Stop){
			elevator.State = ES_IDLE
		} else {
			elevator.State = ES_MOVING
		}
	}
}


func GetElevatorStatus() config.Elevator {
	fmt.Println("-----ElevatorStatus------")
	fmt.Println("floor: ",elevator.Floor)
	fmt.Println("state: ",elevator.State)
	fmt.Println("dir: ",elevator.Direction)

	return elevator
}
