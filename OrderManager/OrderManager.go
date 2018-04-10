package OrderManager

import (
      "../elevio"
      //"log"
      //"io/ioutil"
      //"os"
      //"time"
      //"../Fsm"
)

const NumFloors = 4
const NumButtonsTypes = 3
var ExecuteOrders[4][3] bool


func OrderManager()  {

}


func saveToDisk(filname string) error{

	data, err := jason.Marshal( []byte, error )
		if err != nil{
			log.Println("Eroor: Failed to backup")
			return err
		}

        //func WriteFile(filename string, data []byte, perm os.FileMode) error
		if err := ioutil.WriteFile(filename, data , 0644); err != nil { // writes to file and checks for returned error
			log.Println("Error: Failed to backup")
			return err
		}
		return nil
}

func loadFromDisk(filename string) error { //func Stat(name string) (FileInfo, error)

    var queue

    if _, err := os.State(filename); err == nil {
        data, err := ioutil.ReadFile(filename)
        if err != nil{
        log.Println("Error: Failed to read from backup")
            return err
    }

    if err := jason.Unmarshal(data,queue); err != nil {
        log.Println("Error: failed to Unmarshal")
    }

    }

    return nil

}

func AddOrder(buttonPress elevio.ButtonEvent)  (bool, bool) {

    //floor := <- bt.Floor

    if (ExecuteOrders[buttonPress.Floor][buttonPress.Button] == false){
        ExecuteOrders[buttonPress.Floor][buttonPress.Button] = true
        return true, true
    }

    return false, false
}


//lytter fra chan i main og oppdaterer etasje
func FloorUpdate(floor chan int)  {

}

//lytter fra chan i main og oppdaterer retning
func DirectionUpdate(direction chan elevio.MotorDirection){

}

//??
func LampUpdate()  {

}

func IsElevatorAlive()  bool {

}
