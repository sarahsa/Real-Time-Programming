
package config

const NumFloors = 4
const NumButtonsTypes = 3


type Queue struct{
    matrix [NumButtonsTypes][NumFloors]bool
}

var executeOrders Queue
