package backUp

import (
	"fmt"
	"../config"
	"../elevio"
	"log"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

func SaveToDisk(buttonPress elevio.ButtonEvent, cabOrders[config.N_FLOORS] bool){

	file, err := os.Create("backUp.txt")
	fmt.Println("backup created")
	if err != nil {
		log.Fatal("Cannot create file", err)
	}

	defer file.Close()

	for f := 0; f < config.N_FLOORS; f++ {
		order := strconv.FormatBool(cabOrders[f])
		_ , err = file.WriteString(order)
		_ , err = file.WriteString(" ")
			fmt.Println("writing to file")
			if err != nil {
				log.Fatal("Failed to backup", err)
			}
	}

	//save changes
	err = file.Sync()
	if err != nil{ log.Fatal("Cannot create file", err) }
}

func LoadFromDisk(e config.Elevator)config.Elevator{

	data, err := ioutil.ReadFile("backUp.txt")
	if err != nil {
		log.Fatal("Failed to read from backup", err)
	}
	//fmt.Println("fra load :", string(data))

	 backUpOrders := string(data) // srting
	 backUpOrdersList := strings.Split(backUpOrders, " ") // liste med string

	 for f := 0; f < config.N_FLOORS; f++ {
			 //order, _ := strconv.Atoi(string(buf[f]))
		 //fmt.Println(backUpOrders[f])
		 //fmt.Println("assigning orders from disk: ", backUpOrders)
		 fmt.Println("backUpOrdersList: ",backUpOrdersList[f])
		 order, _ := strconv.ParseBool(backUpOrdersList[f])
		 e.AssignedRequests[f][config.BT_CAB] = order
	 }
	fmt.Println("Loaded from disk array",e.AssignedRequests)
	return e
}

func intToBool(i int)bool  {
	if i == 1 {
		return true
	} else {
		return false
	}
}
