package main 


import(
	"./bcast"
	"fmt"
)

type ButtonPressPacket struct{
	Sender int,
	Floor int,
	Button int
}


func main() {
	
	ButtonPacketTransmit := make(chan ButtonPressPacket)
	ButtonPacketReceive := make(chan ButtonPressPacket)

	go bcast.Transmitter(23232, ButtonPacketTransmit)
	go bcast.Receiver(23232, ButtonPacketReceive)

}