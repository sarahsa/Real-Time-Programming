package sync

import (
	//"../network/network/peers"
	//"../network/network/bcast"
	//"../elevio"

	"time"

	"../config"
	//"fmt"
)

func SendElevatorUpdate(elevator config.Elevator,
	UpdatedElevatorStatus chan config.Elevator,
	ElevatorPacketTrans chan config.ElevatorStatusPacket,
	myID string) {
	for {
		select {
		case updatedElevator := <-UpdatedElevatorStatus:
			elevator = updatedElevator
			//broadcast
			ElevatorPacketTrans <- config.ElevatorStatusPacket{myID, elevator}
		default:
			//Not sure if this works. The channel might lock the code here.
			<-time.After(time.Millisecond * 2000)
			//broadcast
			ElevatorPacketTrans <- config.ElevatorStatusPacket{myID, elevator}

		}
	}
}

/*
func SyncAllLights(Lights config.LightInfoPacket,
	UpdateLightInfoPacket chan config.LightInfoPacket,
	LightPacketTrans chan config.LightInfoPacket) {

	for {
		select {
		case updateLight := <-UpdateLightInfoPacket:
			Lights = updateLight
			fmt.Println("Lights.Button: ", Lights.Button)
			fmt.Println("Lights.Status: ", Lights.Status)
			fmt.Println("Sending lightinfo in update case")
			//LightPacketTrans <- config.LightInfoPacket{Lights.Button, Lights.Status}

		default:
			<-time.After(time.Millisecond * 300)
			//LightPacketTrans <- config.LightInfoPacket{Lights.Button, Lights.Status}

		}
	}

}*/
