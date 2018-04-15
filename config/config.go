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

	//var LocalQueue[N_FLOORS][N_BUTTONS]bool
)

type Elevator struct {
	//ID string
	Floor            int
	State            int
	Direction        elevio.MotorDirection
	AssignedRequests [N_FLOORS][N_BUTTONS]bool
	LightMatrix      [N_FLOORS][N_BUTTONS - 1]bool
	//request ButtonEvent
}
type ElevatorStatusPacket struct {
	ID             string
	ElevatorStatus Elevator
}

type OrderPacket struct {
	Executer string
	//Floor int
	Button elevio.ButtonEvent
	//OrderStatus Status
}

type AcknowledgmentPacket struct {
	Sender   string
	Executer string
	Button   elevio.ButtonEvent
	//OrderStatus Status
}

type OrderMatrix struct {
	AssignedOrders [N_FLOORS][N_BUTTONS]bool
}

type LightInfoPacket struct {
	Button elevio.ButtonEvent
	Status bool
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
