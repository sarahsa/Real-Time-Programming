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
			<-time.After(time.Millisecond * 200)
			//broadcast
			ElevatorPacketTrans <- config.ElevatorStatusPacket{myID, elevator}

		}
	}
}

func SyncLight(Lights config.LightInfo,
	LightInfo chan config.LightInfo,
	LightPacketTrans chan config.LightInfo,
	LightPacketRecv chan config.LightInfo) {

	for {
		select {
		case LightInfo := <-LightPacketRecv:
			Lights = LightInfo
			LightPacketTrans <- Lights

		default:
			<-time.After(time.Millisecond * 200)
			LightPacketTrans <- Lights

		}
	}

}
