package main

import (
        "../elevio"
        "../Fsm"
        //"flag"
        //"../network/network/bcast"
        //"fmt"
        //"../OrderManager"

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


type ButtonPressPacket struct{
	Sender string
	Floor int
	Button int
}

//var ExecuteOrders[4][3] int

/*
func main()  {

    idPtr := flag.String("id", "no id", "elevatro id")
    hwPortPtr := flag.String("hwport", "15657", "select w port hw runs on")
    flag.Parse()
    myID := *idPtr
    hwPort := *hwPortPtr

    elevio.Init(":"+hwPort, 4)

    ButtonPacketTrans := make(chan ButtonPressPacket)
    ButtonPacketRecv := make(chan ButtonPressPacket)
    ButtonPress := make(chan elevio.ButtonEvent)

    go bcast.Transmitter(23232, ButtonPacketTrans)
    go bcast.Receiver(23232, ButtonPacketRecv)

    go elevio.PollHallButtons(ButtonPress)

    for{
        select{
        case buttonPress := <- ButtonPress:
            fmt.Println("Button press at " + myID)
			fmt.Println(buttonPress)

            orderAccepted, changeMade := OrderManager.AddOrder(buttonPress)

            if changeMade {
				ButtonPacketTrans <- ButtonPressPacket{myID, buttonPress.Floor, int(buttonPress.Button)}
			}

			if orderAccepted {
				fmt.OnButtonPress(buttonPress)
			}
        case recvPacket := <-ButtonPacketRecv:
			fmt.Println("Received from " + recvPacket.Sender)
			fmt.Println(recvPacket.Floor, " ", recvPacket.Button)
		}

        }
}


func addOrder(button chan elevio.ButtonEvent)  bool, bool {

    if (ExecuteOrders[button.Floor][button.Button] == false){
        ExecuteOrders[button.Floor][button.Button] = true
        return true, true
    }

    return false, false
}*/
