
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
    floor int
    MotorDirection dir
    requests[N_FLOORS][N_BUTTONS] int
    state int
    ID int
    request ButtonEvent
}

type Queue struct{
    matrix [NumButtonsTypes][NumFloors]bool
}

var executeOrders Queue
