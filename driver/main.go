package main

import (
        "../elevio"
        "../Fsm"
        //"fmt"
        )

func main(){

    numFloors := 4

    elevio.Init("localhost:15657", numFloors)
    

    ch_orders := make(chan elevio.ButtonEvent)
    ch_floors  := make(chan int)
    ch_doorTimeout := make(chan bool)


    go elevio.PollButtons(ch_orders)
    go elevio.PollFloorSensor(ch_floors)

    go Fsm.Fsm(ch_orders, ch_floors, ch_doorTimeout) 


    select{}
}
