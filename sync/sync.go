package sync

import(
  //"../network/network/peers"
  //"../network/network/bcast"
  //"../elevio"
  "../config"
  "time"
  //"fmt"

)
/*
func SyncAllElevatorStatus(updatedElevator chan config.Elevator){
  for{
    select{
      //In this case it will receive the elevatorobjects with updated info,
      //and must update the local info it already has gathered (maps).
      //It also needs to check if it has received any new orders by comparing
      //the local queue with the received order matrix.
      case msg :=<- networkChannel:
        handleMsg(msg)
        //In this case it will send the Elevator object into
        //the assigned channel (networkChannel) every 20th millisecond
      case <- time.After(time.Milllisecond * 20):
          networkChannel <- stateInfo
    }
  }
}
*/
func SendElevatorUpdate(elevator config.Elevator,
                        UpdatedElevatorStatus chan config.Elevator,
                        ElevatorPacketTrans chan config.ElevatorStatusPacket,
                        myID string){
  for{
    select{
    case updatedElevator := <- UpdatedElevatorStatus:
      elevator = updatedElevator
      //broadcast
      ElevatorPacketTrans <- config.ElevatorStatusPacket{myID, elevator}
    default:
      //Not sure if this works. The channel might lock the code here.
    <- time.After(time.Millisecond * 2000)
    //broadcast
    ElevatorPacketTrans <- config.ElevatorStatusPacket{myID, elevator}

    }
  }
}
