
package config
// def "config"

import "../elevio"


const (

    N_ELEVATORS = 3
    N_FLOORS = 4
    N_BUTTONS = 3

    BT_HallUp = 0
    BT_HallDown = 1
    BT_CAB = 2

)

type Elevator struct {
    ID string
    floor int
    state int
    direction elevio.MotorDirection
    requests[N_FLOORS][N_BUTTONS] bool
    //request ButtonEvent
}

type Queue struct{
    matrix [N_BUTTONS][N_FLOORS]bool
}

var executeOrders Queue
