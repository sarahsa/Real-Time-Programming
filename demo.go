package main

import
(
	"./elevio"
	"fmt"
)

func main(){

	elevio.Init("localhost:15657", 4)

	buttonPressedCh := make(chan elevio.ButtonEvent)
	floorReachedCh := make(chan int)

	go elevio.PollButtons(buttonPressedCh)
	go elevio.PollFloorSensor(floorReachedCh)

	var orderMatrix [4][3]bool

	for{
		select{
			case msg := <- buttonPressedCh:
				// Add to orders
				orderMatrix[msg.Floor][msg.Button] = true
				fmt.Println(orderMatrix)

			case msg := <- floorReachedCh:
				fmt.Println(msg)
		}
	}

}