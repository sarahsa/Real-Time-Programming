
# Finite State Machine

This module uses states from config. 

|  States   | Event        |
| --------- |---------     |
| ES_INIT   | newOrder     |
| ES_IDLE   | reachedFloor |
|ES_MOVING  | doortimer    |      
|ES_DOOROPEN| motortimer   |
|ES_STUCK   |              |   



**ES_INIT** - _Initializing elevator to drive up to nearest floor and stay there with door closed_  
**ES_IDLE** - _Waiting for orders at a floor with foor closed._  
**ES_MOVING** - _Elevator is moving.  It is either between floors or passing a floor._  
**ES_DOOROPEN** - _The elevator has reached a floor with order. The door is opend for three seconds._  
**ES_STUCK** - _The elevator is not moving_  

