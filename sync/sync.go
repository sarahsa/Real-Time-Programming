package sync

import (
	//"../network/network/peers"
	//"../network/network/bcast"
	"../elevio"

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
			//Not sure if this works. The channel might lock the code here.
		case <-time.After(time.Millisecond * 200):
			//broadcast
			ElevatorPacketTrans <- config.ElevatorStatusPacket{myID, elevator}

		}
	}
}

func SyncAllLights(OrderRegistered [config.N_FLOORS][config.N_BUTTONS - 1]bool) {
	for f := 0; f < config.N_FLOORS; f++ {
		for b := 0; b < config.N_BUTTONS-1; b++ {
			elevio.SetButtonLamp(elevio.ButtonType(b), f, OrderRegistered[f][b])
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
