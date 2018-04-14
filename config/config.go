package config

// def "config"

import "../elevio"

const (
	N_ELEVATORS = 3
	N_FLOORS    = 4
	N_BUTTONS   = 3

	BT_HallUp   = 0
	BT_HallDown = 1
	BT_CAB      = 2
)

type Elevator struct {
	//ID string
	Floor            int
	State            int
	Direction        elevio.MotorDirection
	AssignedRequests [N_FLOORS][N_BUTTONS]bool
	//request ButtonEvent
}
type ElevatorStatusPacket struct {
	ID             string
	ElevatorStatus Elevator
}

/*type ButtonPressPacket struct{
	Sender string
	Floor int
	Button elevio.ButtonType
  //StatusOrder Status
}*/

type ButtonPressPacket struct {
	Executer string
	//Floor int
	Button elevio.ButtonEvent
	//StatusOrder Status
}

type Status int

const (
	OrderNotAssigned Status = 0
	OrderAssigned           = 1
	OrderExecuted           = 2
)

type Queue struct {
	matrix [N_BUTTONS][N_FLOORS]bool
}

var executeOrders Queue
