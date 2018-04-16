
# TTK4145-Sanntidsprogrammering

### Language: Go

### 
We decided to solve the project by having a shifting "Master" with the use of UDP Broadcasting. In order to do so, all the participating elevators have access to each others position, moving direction, current state and the assigned hall orders. 
The hall orders are assigned by the elevator which receives the button on its elevator panel, and broadcasted to the other elevators.

In order for an order to be executed, one needs to be sure that at least one of the other elevators have ackowledged that there has been registered a new order and that the order has been assigned. 

### Modules
- FSM

