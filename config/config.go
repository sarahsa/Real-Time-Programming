
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
    ID int
    floor int
    requests[N_FLOORS][N_BUTTONS] int
    state int
    MotorDirection elevoi.MotorDirection
    //request ButtonEvent
}

type Queue struct{
    matrix [NumButtonsTypes][NumFloors]bool
}

var executeOrders Queue
