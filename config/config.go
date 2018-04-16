package config

import (
	"time"
	"../elevio"
)
const (
	N_ELEVATORS = 3
	N_FLOORS    = 4
	N_BUTTONS   = 3

	CH_BUFFERSIZE = 10

	DOOR_OPEN_TIME time.Duration = 3000
	TRAVEL_TIME    time.Duration = 2500

	ES_INIT     = 0 
	ES_IDLE     = 1
	ES_MOVING   = 2
	ES_DOOROPEN = 3
	ES_STUCK    = 4
)

type Elevator struct {
	Floor            int
	State            int
	Direction        elevio.MotorDirection
	AssignedRequests [N_FLOORS][N_BUTTONS]bool
	LightMatrix      [N_FLOORS][N_BUTTONS - 1]bool
}

type ElevatorStatusPacket struct {
	ID             string
	ElevatorStatus Elevator
}

type OrderPacket struct {
	Executer string
	Button elevio.ButtonEvent
}

type AcknowledgmentPacket struct {
	Sender   string
	Executer string
	Button   elevio.ButtonEvent
}

type OrderMatrix struct {
	AssignedOrders [N_FLOORS][N_BUTTONS]bool
}

type ReceivedAck struct {
	Button elevio.ButtonEvent
	Status bool
}

type Queue struct {
	matrix [N_BUTTONS][N_FLOORS]bool
}

var executeOrders Queue
