
package config



const (
    ELEVATORS = 3
    FLOORS = 4
    BUTTONS = 3

    BT_HallUp = 0
    BT_HallDown = 1
    BT_CAB = 2

    

)

type Queue struct{
    matrix [NumButtonsTypes][NumFloors]bool
}

var executeOrders Queue
