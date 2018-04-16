package backUp

import (
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
	if err != nil {
		log.Fatal("Cannot create file", err)
	}

	defer file.Close()

	for f := 0; f < config.N_FLOORS; f++ {
		order := strconv.FormatBool(cabOrders[f])
		_ , err = file.WriteString(order)
		_ , err = file.WriteString(" ")
			if err != nil {
				log.Fatal("Failed to backup", err)
			}
	}
	err = file.Sync()
	if err != nil{ log.Fatal("Cannot create file", err) }
}

func LoadFromDisk(e config.Elevator)config.Elevator{

	data, err := ioutil.ReadFile("backUp.txt")
	if err != nil {
		log.Fatal("Failed to read from backup", err)
	}

	 backUpOrders := string(data)
	 backUpOrdersList := strings.Split(backUpOrders, " ")

	 for f := 0; f < config.N_FLOORS; f++ {
		 order, _ := strconv.ParseBool(backUpOrdersList[f])
		 e.AssignedRequests[f][elevio.BT_Cab] = order
	 }
	return e
}
