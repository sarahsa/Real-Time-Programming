
# OrderManager

* Updated numbers of elevator active on the network or alive
  - if elev1 disconnects form the network:
    - elev2 or elev3 (elevtatir with lowest ID) will reassign elevator1's hall orders between elev2 and elev3
  * if only elevator1 is alive or elev2 and elev3 is disconnected from the network:
    - the elevators will only take cab orders 
* Broadcasts and recives orders between elevators
* Uses CalculateCost() for assigning orders
* Takes backup of caborders 
* Sets buttonlights on when reciving acknowledgment
