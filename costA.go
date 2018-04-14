

func Delegate(elevators map[string]*Elevator, aliveElevators map[string]bool, floor int, button ButtonType) string {
    bestTime := time.Duration(math.MaxInt64)
    bestId   := ""

    fmt.Printf("\n --- Delegating order {f:%v, b:%v} --- \n", floor, button)

    for id, e := range elevators {
        if aliveElevators[id] {
            fmt.Printf("%v:\n", id)
            e2 := e.Dup()
            e2.Requests[floor][button] = 1
            c := TimeToIdle(e2)
            if c < bestTime {
                bestTime = c
                bestId = id
            }
        }
    }
    fmt.Printf(" --- Delegate best: [%v] --- \n\n\n", bestId)
    return bestId
}

func TimeToIdle(e Elevator) time.Duration {

    dur := 0*time.Millisecond

    fmt.Printf("\t  %s%+v%s\n", Color_req, e, Color_reset)

    switch e.Behaviour {
    case EB_Idle:
        e.Dirn = ChooseDirn(e)
        if e.Dirn == MD_Stop { return dur }
    case EB_Moving:
        e.Floor = e.Floor + int(e.Dirn)
        dur += e.Config.TravelTime/2
    case EB_DoorOpen:
        dur -= e.Config.DoorOpenTime/2
    }

    for {
        fmt.Printf("%s%+v, \t| %+v%s\n", Color_req, dur, e, Color_reset)
        if ShouldStop(e) {
            e = ClearRequestsAtCurrentFloor(e, nil)
            dur += e.Config.DoorOpenTime
            e.Dirn = ChooseDirn(e)
            if e.Dirn == MD_Stop {
                fmt.Printf("%sdur: [%+v]%s\n", Color_req, dur, Color_reset)
                return dur
            }
        }
        e.Floor = e.Floor + int(e.Dirn)
        dur += e.Config.TravelTime
    }
}cos
