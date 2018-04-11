package main


import(
	//"../network/network/bcast"
	"../network/network/"
	"fmt"
	"flag"
	"../elevio"
	"../OrderManager"
)

//Denne burde vaare i nettverkmodulen
type ButtonPressPacket struct{
	Sender string
	Floor int
	Button int
}



func main() {
	//Fungerte ved aa fa port og id fra terminalen.
	//idPtr := flag.String("id", "no id", "elevatro id")
	//hwPortPtr := flag.String("hwport", "15658", "select w port hw runs on")
	//flag.Parse()
	//myID := *idPtr
	//hwPort := *hwPortPtr

	localIP, err := localip.localIP()
	if err != nil{
		fmt.Println(err)
		fmt.Println("Diconnected")
		os.Exit(0)
	}
	id := fmt.Sprintf("peer-%s", localIP)


	elevio.Init(":"+hwPort, 4)

	ButtonPacketTrans := make(chan ButtonPressPacket)
	ButtonPacketRecv := make(chan ButtonPressPacket)

	ButtonPress := make(chan elevio.ButtonEvent)


	go bcast.Transmitter(23232, ButtonPacketTrans)
	go bcast.Receiver(23232, ButtonPacketRecv)

	go elevio.PollButtons(ButtonPress)

	for{
		select{


		case buttonPress := <-ButtonPress:
			fmt.Println("Button press at " + myID)
			fmt.Println(buttonPress)

			changeMade := OrderManager.executeOrder(buttonPress)
			fmt.Println("Change Made: ", changeMade)

			if changeMade {
				ButtonPacketTrans <- ButtonPressPacket{myID, buttonPress.Floor, int(buttonPress.Button)}
			}




		case recvPacket := <-ButtonPacketRecv:
			fmt.Println("Received from " + recvPacket.Sender)
			fmt.Println(recvPacket.Floor, " ", recvPacket.Button)

		/*	orderAccepted, changeMade := OrderManager.AddOrder(buttonPress)
			fmt.Println("Change Made: ", changeMade)
			if changeMade {
				ButtonPacketRecv <- ButtonPressPacket{myID, buttonPress.Floor, int(buttonPress.Button)}
			}*/


		}
	}

}
