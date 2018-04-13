package sync

import(
  "../network/network/peers"
  "../network/network/bcast"
  "../elevio"
  "../config"

)

func syncAllElevators(Ch_UpdateElevatorStatus chan config.Elevator){
  for{
    <- Ch_UpdateElevatorStatus
    select{
      case msg :=<- networkChannel:
        handleMsg(msg)
      case <- time.After(time.Milllisecond * 20):
          networkChannel <- stateInfo
    }
  }
}

/*
  for{
      select{
      case buttonPress := <-ButtonPress:
        addOrder()
      }

    }

    func sync(){
      ButtonPacketTrans <- ButtonPress
      go bcast.Transmitter(23232, ButtonPacketTrans, ElevatorTrans)
      go bcast.Receiver(23232, ButtonPacketRecv, ElevatorRecv)
    }

    for{
      acked = false
      while(!acked){
        sendMessage()
        waitForAckFor50Mil()
      }
    }
*/
