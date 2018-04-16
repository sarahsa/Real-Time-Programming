
# TTK4145-Sanntidsprogrammering

### Language: Go

### 
We decided to solve the project by having a shifting "Master" with the use of UDP Broadcasting. In order to do so, all the participating elevators have access to each others position, moving direction, current state and the assigned hall orders. 
The hall orders are assigned by the elevator which receives the button on its elevator panel, and broadcasted to the other elevators.



### Modules

- Order Manager:

The Order Manager can be considered the "brain" of this system. Whenever a new elevator is connected to the network the elevator is added to `activeElevators`, and whenever an elevator looses network it is also removed from this map. This module also has the responsibility of both assigning an order when it is first registered, and reassigning them if an elevator dies with hall orders to be executed. This is done by calculating the *cost for each of the active elevators and assigning the order to the elevator with lowest cost, as this will make the system consisting of `n=3` elevators more efficient than only one single elevator. In order to calculate the cost, this module has to have access to updated information regarding the position of all the participating elevators. Since the order manager has the main responsibility of communicating with the other elevators as well, it can easily access the required information for the cost algorithm. For an order to be executed, one needs to be sure that at least one of the other elevators have ackowledged that there has been registered a new order and that the order has been assigned. When this acknowledgement is received, the order is transmitted to the FSM which executes the order. 

*cost: the function calculating the cost is implemented using [cost function](https://github.com/TTK4145/Project-resources/tree/master/cost_fns).

- FSM

As mentioned above the FSM is regarded the executing part of the system. It receives orders from the order manager, and executes them. The FSM is also using the Elevator I/0 module, in order to interface with the hardware. 

- Network

The network module uses UDP broadcasting and is found [here](https://github.com/TTK4145/Network-go)


- Backup

This module has the responsibility of writing all the received caborders to an external file, so that in a case of an error the orders are not lost. 


- Elevator I/O

This module is only for interfacing with the hardware

