package Fsm

import (
	"fmt"
	"time"
	"../backUp"
	"../config"
	"../elevio"
)

var elevator config.Elevator
var lastDirection elevio.MotorDirection //Kan hende denne er unødvendig
var ch_floors = make(chan int)

func Fsm(ch_assignedOrders chan elevio.ButtonEvent,
	ch_doorTimeout chan bool,
	ch_updateElevatorStatus chan config.Elevator,
	ch_orderIsExecuted chan elevio.ButtonEvent,
	ch_timedOutMotor chan config.OrderMatrix) {

	Init()
	go elevio.PollFloorSensor(ch_floors)

	//creating timers
	doortimer := time.NewTimer(3 * time.Second)
	doortimer.Stop()
	motortimer := time.NewTimer(5 * time.Second)
	motortimer.Stop()

	for {
		//Only for debugging purposes
		switch elevator.State {
		case config.ES_INIT:
			fmt.Println("State: Init")
		case config.ES_IDLE:
			fmt.Println("State: Idle")
		case config.ES_MOVING:
			fmt.Println("State: Moving")
		case config.ES_DOOROPEN:
			fmt.Println("State: DoorOpen")
		case config.ES_STUCK:
			fmt.Println("State: stuck")
		}

		switch elevator.Direction {
		case elevio.MD_Up:
			fmt.Println("Dir: Up")
		case elevio.MD_Down:
			fmt.Println("Dir: Down")
		case elevio.MD_Stop:
			fmt.Println("Dir: Stop")
		}

		switch lastDirection {
		case elevio.MD_Up:
			fmt.Println("LastDir: Up")
		case elevio.MD_Down:
			fmt.Println("LastDir: Down")
		case elevio.MD_Stop:
			fmt.Println("LastDir: Stop")
		}

		select {
		case newOrder := <-ch_assignedOrders:
			switch elevator.State {
			case config.ES_IDLE:
				if newOrder.Floor == elevator.Floor {
					ArrivedAtOrderedFloor(newOrder, ch_orderIsExecuted)
					doortimer.Reset(3 * time.Second)
					ch_updateElevatorStatus <- elevator
				} else {
					addOrder(newOrder)
					//fmt.Println("Orders: %v", elevator.AssignedRequests)
					executeOrder(elevator.Floor, newOrder.Floor)
					elevator.State = config.ES_MOVING
					//fmt.Println("REset motortimer")
					motortimer.Reset(5 * time.Second)
					ch_updateElevatorStatus <- elevator
				}
			case config.ES_DOOROPEN:
				if newOrder.Floor == elevator.Floor {
					doortimer.Reset(3 * time.Second)
					//evt bare continue:
				} else {
					addOrder(newOrder)
					//fallthrough
				}
			default:
				addOrder(newOrder)
				//fmt.Println("Orders: %v", elevator.AssignedRequests)
			}

		case reachedFloor := <-ch_floors:
			elevio.SetFloorIndicator(reachedFloor)
			elevator.Floor = reachedFloor
			//fmt.Println("floor from reachedFloor: ", elevator.Floor)
			switch elevator.State {
			case config.ES_INIT:
				SetDirectionAfterInit(ch_updateElevatorStatus)

			case config.ES_MOVING:
				//fmt.Println("REset motortimer")
				motortimer.Reset(10 * time.Second)
				if reachedFloor == 3 || reachedFloor == 0 {
					fmt.Println("Changing direction due to 0 or 3")
					SwapDirection()
				}
				if CheckOrdersAtFloor(reachedFloor) {
					//Kan lage funksjon av dette:
					UpdateDirection(elevio.MD_Stop)

					if (elevator.AssignedRequests[reachedFloor][elevio.BT_Cab]) && !(elevator.AssignedRequests[reachedFloor][elevio.BT_HallUp] || elevator.AssignedRequests[reachedFloor][elevio.BT_HallDown]) {
						ch_orderIsExecuted <- elevio.ButtonEvent{reachedFloor, elevio.BT_Cab}
					}
					ArrivedAtOrderedFloor(elevio.ButtonEvent{reachedFloor, FromMotorDirectionToButton()}, ch_orderIsExecuted)
					doortimer.Reset(3 * time.Second)
					ch_updateElevatorStatus <- elevator
				} else {
					SwapDirection()
					if CheckOrdersAtFloor(reachedFloor) {
						UpdateDirection(elevio.MD_Stop)
						ClearOrdersAtCurrentFloor(reachedFloor)
						ArrivedAtOrderedFloor(elevio.ButtonEvent{reachedFloor, FromMotorDirectionToButton()}, ch_orderIsExecuted)
						ch_updateElevatorStatus <- elevator
					}
				}
			case config.ES_STUCK:
				elevio.SetMotorDirection(elevio.MD_Stop)
				if CheckOrdersAtFloor(reachedFloor) {
					UpdateDirection(elevio.MD_Stop)
				}
				if CheckIfAnyOrders() {
					elevator.State = config.ES_MOVING
					elevio.SetMotorDirection(elevator.Direction)
				} else {
					elevator.State = config.ES_IDLE
				}
			}

		case <-doortimer.C:
			//fmt.Println("DoorTimeout")
			elevio.SetDoorOpenLamp(false)
			if CheckIfAnyOrders() {
				elevator.Direction = lastDirection
				elevator.State = config.ES_MOVING
				motortimer.Reset(5 * time.Second)

				//Koden under her kan forenkles:
				if CheckUpcomingFloors(elevator) {
					//fmt.Println("inside doortimer - checkupcomingfloor")
					elevator.Direction = lastDirection
					elevio.SetMotorDirection(elevator.Direction)
				} else {
					SwapDirection()
					//fmt.Println("Inside doortimer - else statement")
					if CheckUpcomingFloors(elevator) {
						//fmt.Println("dir = ", elevator.Direction)
						elevio.SetMotorDirection(elevator.Direction)
					} else {
						elevator.Direction = lastDirection
						elevio.SetMotorDirection(elevator.Direction)
					}
				}
			} else {
				elevator.State = config.ES_IDLE
				motortimer.Stop()
			}
		case <-motortimer.C:
			//fmt.Println("Motor timed out")
			ch_timedOutMotor <- config.OrderMatrix{elevator.AssignedRequests}
			clearHallOrders()
			elevator.State = config.ES_STUCK
		}
	}
}

func Init() {
	fmt.Println("Initializing...")
	elevio.SetDoorOpenLamp(false)
	fmt.Println("doorclosed")
	elevator.State = config.ES_INIT

	for f := 0; f < 4; f++ {
		for b := elevio.ButtonType(0); b < 3; b++ {
			elevio.SetButtonLamp(b, f, false)
		}
	}
	elevator = backUp.LoadFromDisk(elevator)
	for f := 0; f < config.N_FLOORS; f++ {
		if elevator.AssignedRequests[f][elevio.BT_Cab] == true {
			elevio.SetButtonLamp(elevio.BT_Cab,f,true)
		}
	}
	if elevio.GetFloor() == -1 {
		elevio.SetMotorDirection(elevio.MD_Up)
		elevator.Direction = elevio.MD_Up
	} else {
		elevator.Floor = elevio.GetFloor()
	}
}

func SetDirectionAfterInit(ch_updateElevatorStatus chan config.Elevator){
	if IsOrderAbove(elevator){
		elevator.Direction = elevio.MD_Up
		elevator.State = config.ES_MOVING
		elevio.SetMotorDirection(elevator.Direction)

	} else if IsOrderBelow(elevator){
		elevator.Direction = elevio.MD_Down
		elevator.State = config.ES_MOVING
		elevio.SetMotorDirection(elevator.Direction)
	} else {
		elevio.SetMotorDirection(elevio.MD_Stop)
		lastDirection = elevator.Direction
		fmt.Println("lastDirection is set to: ", lastDirection)
		elevator.Direction = elevio.MD_Stop
		elevator.State = config.ES_IDLE
		ch_updateElevatorStatus <- elevator
	}

}

func clearHallOrders() {
	for f := 0; f < config.N_FLOORS; f++ {
		for b := 0; b < config.N_BUTTONS-1; b++ {
			elevator.AssignedRequests[f][b] = false
		}
	}
}

func addOrder(pressedButton elevio.ButtonEvent) {
	fmt.Println("Er inn i addOrder")
	if elevator.AssignedRequests[pressedButton.Floor][pressedButton.Button] == false {
		elevator.AssignedRequests[pressedButton.Floor][pressedButton.Button] = true
		elevio.SetButtonLamp(pressedButton.Button, pressedButton.Floor, true)

	}
}

func SwapDirection() {
	fmt.Println("Changing direction ")
	if elevator.Direction == elevio.MD_Up || elevator.Floor == 3 {
		elevator.Direction = elevio.MD_Down
	} else if elevator.Direction == elevio.MD_Down || elevator.Floor == 0 {
		elevator.Direction = elevio.MD_Up
	}
}
func UpdateDirection(newDir elevio.MotorDirection) {
	lastDirection = elevator.Direction
	fmt.Println("lastDirection is set to: ", lastDirection)
	elevator.Direction = newDir
	elevio.SetMotorDirection(elevator.Direction)

}

func executeOrder(floor int, targetFloor int) {

	if floor < targetFloor {
		elevator.Direction = elevio.MD_Up

	} else if floor > targetFloor {
		elevator.Direction = elevio.MD_Down

	} else { //elevator.Floor == targetFloor
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

func CheckIfAnyOrders() bool {
	for f := 0; f < 4; f++ {
		for b := 0; b < 3; b++ {
			if elevator.AssignedRequests[f][b] == true {
				fmt.Println("It has more orders to execute")
				return true
			}
		}
	}
	return false
}


func CheckIfAnyHallOrders()  {
	//
}

func CheckIfAnyCabOrders(){
	//
}

func CheckOrdersAtFloor(floor int) bool {
	switch elevator.Direction {
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

func ArrivedAtOrderedFloor(newOrder elevio.ButtonEvent, ch_orderIsExecuted chan elevio.ButtonEvent){
	ClearOrdersAtCurrentFloor(elevator.Floor)
	ch_orderIsExecuted <- newOrder
	elevator.State = config.ES_DOOROPEN
	elevio.SetDoorOpenLamp(true)

}

func CheckUpcomingFloors(e config.Elevator) bool {
	switch elevator.Direction {
	case elevio.MD_Up:
		return IsOrderAbove(e)
	case elevio.MD_Down:
		return IsOrderBelow(e)
	}
	return false
}

func IsOrderAbove(e config.Elevator) bool {
	if e.Floor == 3 {
		return false
	}
	for f := e.Floor + 1; f < 4; f++ {
		for b := 0; b < config.N_BUTTONS; b++ {
			if e.AssignedRequests[f][b] {
				return true
			}
		}
	}
	return false
}

func IsOrderBelow(e config.Elevator) bool {
	if e.Floor == 0 {
		return false
	}
	for f := 0; f < e.Floor; f++ {
		for b := 0; b < config.N_BUTTONS; b++ {
			if e.AssignedRequests[f][b] {
				return true
			}
		}
	}
	return false
}

func ClearOrdersAtCurrentFloor(floor int) {
	//fmt.Println("Er inne i ClearOrdersAtCurrentFloor")

	elevator.AssignedRequests[floor][elevio.BT_Cab] = false
	elevio.SetButtonLamp(elevio.BT_Cab, floor, false)
/*
	if elevator.AssignedRequests[floor][elevio.BT_HallUp] == true {
		elevator.AssignedRequests[floor][elevio.BT_HallUp] = false
	} else if elevator.AssignedRequests[floor][elevio.BT_HallDown]==true {
		elevator.AssignedRequests[floor][elevio.BT_HallDown] = false
	}*/

	switch elevator.Direction {
	case elevio.MD_Up:
		elevator.AssignedRequests[floor][elevio.BT_HallUp] = false
		elevio.SetButtonLamp(elevio.BT_HallUp, floor, false)
		fmt.Println("ClearOrders: Case MD_UP")
		if !IsOrderAbove(elevator) {
			elevator.AssignedRequests[floor][elevio.BT_HallDown] = false
			elevio.SetButtonLamp(elevio.BT_HallDown, floor, false)
		}
	case elevio.MD_Down:
		elevator.AssignedRequests[floor][elevio.BT_HallDown] = false
		elevio.SetButtonLamp(elevio.BT_HallDown, floor, false)
		fmt.Println("ClearOrders: Case MD_Down")
		if !IsOrderBelow(elevator) {
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

func FromMotorDirectionToButton() elevio.ButtonType {
	if lastDirection == elevio.MD_Up {
		return elevio.BT_HallUp
	} else {
		return elevio.BT_HallDown
	}
}

func DoorTimeout() {
	fmt.Println("Er inne i DoorTimeout")

	if elevator.State == config.ES_DOOROPEN {
		//direction = ChooseDirection(currentFloor)
		elevio.SetDoorOpenLamp(false)
		elevio.SetMotorDirection(elevator.Direction)

		if elevator.Direction == elevio.MD_Stop {
			elevator.State = config.ES_IDLE
		} else {
			elevator.State = config.ES_MOVING
		}
	}
}

func GetElevatorStatus() config.Elevator {
	fmt.Println("-----ElevatorStatus------")
	fmt.Println("floor: ", elevator.Floor)
	fmt.Println("state: ", elevator.State)
	fmt.Println("dir: ", elevator.Direction)

	return elevator

}
