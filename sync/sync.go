package sync

import (
	"../elevio"
	"time"
	"../config"
)

func SendElevatorUpdate(elevator config.Elevator,
	UpdatedElevatorStatus chan config.Elevator,
	ElevatorPacketTrans chan config.ElevatorStatusPacket,
	myID string) {
	for {
		select {
		case updatedElevator := <-UpdatedElevatorStatus:
			elevator = updatedElevator
			ElevatorPacketTrans <- config.ElevatorStatusPacket{myID, elevator}
		case <-time.After(time.Millisecond * 2000):
			ElevatorPacketTrans <- config.ElevatorStatusPacket{myID, elevator}

		}
	}
}

func SyncAllLights(AssignedRequests [config.N_FLOORS][config.N_BUTTONS]bool) {
	for f := 0; f < config.N_FLOORS; f++ {
		for b := 0; b < config.N_BUTTONS-1; b++ {
			elevio.SetButtonLamp(elevio.ButtonType(b), f, AssignedRequests[f][b])
		}
	}
}
