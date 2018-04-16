package Fsm

import (
	"time"
	"../backUp"
	"../config"
	"../elevio"
)

const (
	ES_INIT     = 0
	ES_IDLE     = 1
	ES_MOVING   = 2
	ES_DOOROPEN = 3
	ES_STUCK    = 4
)

var elevator config.Elevator
var lastDirection elevio.MotorDirection
var ch_floors = make(chan int)

func Fsm(ch_assignedOrders chan elevio.ButtonEvent,
	ch_doorTimeout chan bool,
	ch_updateElevatorStatus chan config.Elevator,
	ch_orderIsExecuted chan elevio.ButtonEvent,
	ch_timedOutMotor chan config.OrderMatrix)  {

	Init()
	go elevio.PollFloorSensor(ch_floors)
	doortimer := time.NewTimer(3 * time.Second)
	doortimer.Stop()
	motortimer := time.NewTimer(5 * time.Second)
	motortimer.Stop()
	for {
		select {
		case newOrder := <-ch_assignedOrders:
			switch elevator.State {
			case ES_IDLE:
				if newOrder.Floor == elevator.Floor {
					ClearOrdersAtCurrentFloor(elevator.Floor)
					ch_orderIsExecuted <- newOrder
					elevator.State = ES_DOOROPEN
					elevio.SetDoorOpenLamp(true)
					doortimer.Reset(3 * time.Second)
					ch_updateElevatorStatus <- elevator
				} else {
					AddOrder(newOrder)
					ExecuteOrder(elevator.Floor, newOrder.Floor)
					elevator.State = ES_MOVING
					motortimer.Reset(5 * time.Second)
					ch_updateElevatorStatus <- elevator
				}
			case ES_DOOROPEN:
				if newOrder.Floor == elevator.Floor {
					doortimer.Reset(3 * time.Second)
				} else {
					AddOrder(newOrder)
				}
			default:
				AddOrder(newOrder)
			}

		case reachedFloor := <-ch_floors:
			elevio.SetFloorIndicator(reachedFloor)
			elevator.Floor = reachedFloor
			switch elevator.State {
			case ES_INIT:
				if IsOrderAbove(elevator) {
					elevator.Direction = elevio.MD_Up
					elevator.State = ES_MOVING
					elevio.SetMotorDirection(elevator.Direction)
				} else if IsOrderBelow(elevator) {
					elevator.Direction = elevio.MD_Down
					elevator.State = ES_MOVING
					elevio.SetMotorDirection(elevator.Direction)
				}else{
					elevio.SetMotorDirection(elevio.MD_Stop)
					lastDirection = elevator.Direction
					elevator.Direction = elevio.MD_Stop
					elevator.State = ES_IDLE
					ch_updateElevatorStatus <- elevator
				}

			case ES_MOVING:
				motortimer.Reset(10 * time.Second)
				if reachedFloor == 3 || reachedFloor == 0 {
					ChangeDirection()
				}
				if CheckOrdersAtFloor(reachedFloor) {
					lastDirection = elevator.Direction
					elevator.Direction = elevio.MD_Stop
					elevio.SetMotorDirection(elevator.Direction)
					if (elevator.AssignedRequests[reachedFloor][elevio.BT_Cab]) && !(elevator.AssignedRequests[reachedFloor][elevio.BT_HallUp] || elevator.AssignedRequests[reachedFloor][elevio.BT_HallDown]) {
						ch_orderIsExecuted <- elevio.ButtonEvent{reachedFloor, elevio.BT_Cab}
					}

					ch_orderIsExecuted <- elevio.ButtonEvent{reachedFloor, FromMotorDirectionToButton()}

					ClearOrdersAtCurrentFloor(elevator.Floor)
					elevio.SetDoorOpenLamp(true)
					doortimer.Reset(3 * time.Second)
					elevator.State = ES_DOOROPEN
					ch_updateElevatorStatus <- elevator
				} else {
					ChangeDirection()
					if CheckOrdersAtFloor(reachedFloor) {
						lastDirection = elevator.Direction
						elevator.Direction = elevio.MD_Stop
						elevio.SetMotorDirection(elevator.Direction)
						ClearOrdersAtCurrentFloor(elevator.Floor)
						ch_orderIsExecuted <- elevio.ButtonEvent{reachedFloor, FromMotorDirectionToButton()}
						elevio.SetDoorOpenLamp(true)
						doortimer.Reset(3 * time.Second)
						elevator.State = ES_DOOROPEN
						ch_updateElevatorStatus <- elevator
					}
				}
			case ES_STUCK:
				elevio.SetMotorDirection(elevio.MD_Stop)
				if CheckOrdersAtFloor(reachedFloor) {
					lastDirection = elevator.Direction
					elevator.Direction = elevio.MD_Stop
					elevio.SetMotorDirection(elevio.MD_Stop)

				}
				if CheckIfAnyOrders() {
					elevator.State = ES_MOVING
					elevio.SetMotorDirection(elevator.Direction)
				} else {
					elevator.State = ES_IDLE
				}
			}

		case <-doortimer.C:
			elevio.SetDoorOpenLamp(false)
			if CheckIfAnyOrders() {
				elevator.Direction = lastDirection
				elevator.State = ES_MOVING
				motortimer.Reset(5 * time.Second)
				if CheckUpcomingFloors(elevator) {
					elevator.Direction = lastDirection
					elevio.SetMotorDirection(elevator.Direction)
				} else {
					ChangeDirection()
					if CheckUpcomingFloors(elevator) {
						elevio.SetMotorDirection(elevator.Direction)
					} else {
						elevator.Direction = lastDirection
						elevio.SetMotorDirection(elevator.Direction)
					}
				}
			} else {
				elevator.State = ES_IDLE
				motortimer.Stop()
			}
		case <-motortimer.C:
			ch_timedOutMotor <- config.OrderMatrix{elevator.AssignedRequests}
			ClearHallOrders()
			elevator.State = ES_STUCK
	}
	}
}

func Init() {

	elevio.SetDoorOpenLamp(false)
	elevator.State = ES_INIT

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

func ClearHallOrders() {
	for f := 0; f < config.N_FLOORS; f++ {
		for b := 0; b < config.N_BUTTONS-1; b++ {
			elevator.AssignedRequests[f][b] = false
		}
	}
}

func AddOrder(pressedButton elevio.ButtonEvent) {
	if elevator.AssignedRequests[pressedButton.Floor][pressedButton.Button] == false {
		elevator.AssignedRequests[pressedButton.Floor][pressedButton.Button] = true
		elevio.SetButtonLamp(pressedButton.Button, pressedButton.Floor, true)

	}
}

func ChangeDirection() {
	if elevator.Direction == elevio.MD_Up || elevator.Floor == 3 {
		elevator.Direction = elevio.MD_Down
	} else if elevator.Direction == elevio.MD_Down || elevator.Floor == 0 {
		elevator.Direction = elevio.MD_Up
	}
}

func ExecuteOrder(floor int, targetFloor int) {

	if floor < targetFloor {
		elevator.Direction = elevio.MD_Up
	} else if floor > targetFloor {
		elevator.Direction = elevio.MD_Down
	} else {
		lastDirection = elevator.Direction
		elevator.Direction = elevio.MD_Stop
		ClearOrdersAtCurrentFloor(floor)
		elevio.SetDoorOpenLamp(true)
		elevator.Direction = lastDirection
	}
	elevio.SetMotorDirection(elevator.Direction)
}

func CheckIfAnyOrders() bool {
	for f := 0; f < 4; f++ {
		for b := 0; b < 3; b++ {
			if elevator.AssignedRequests[f][b] == true {
				return true
			}
		}
	}
	return false
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
		return false
	}
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
	elevator.AssignedRequests[floor][elevio.BT_Cab] = false
	elevio.SetButtonLamp(elevio.BT_Cab, floor, false)
	switch elevator.Direction {
	case elevio.MD_Up:
		elevator.AssignedRequests[floor][elevio.BT_HallUp] = false
		elevio.SetButtonLamp(elevio.BT_HallUp, floor, false)
		if !IsOrderAbove(elevator) {
			elevator.AssignedRequests[floor][elevio.BT_HallDown] = false
			elevio.SetButtonLamp(elevio.BT_HallDown, floor, false)
		}
	case elevio.MD_Down:
		elevator.AssignedRequests[floor][elevio.BT_HallDown] = false
		elevio.SetButtonLamp(elevio.BT_HallDown, floor, false)
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
	if elevator.State == ES_DOOROPEN {
		elevio.SetDoorOpenLamp(false)
		elevio.SetMotorDirection(elevator.Direction)

		if elevator.Direction == elevio.MD_Stop {
			elevator.State = ES_IDLE
		} else {
			elevator.State = ES_MOVING
		}
	}
}

func GetElevatorStatus() config.Elevator {
	return elevator
}
