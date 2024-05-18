# Can-Coder
This repository contains a can interface which enables to connect to a car's canbus.

Goal is to convert the can frames into a human readable format and forward it in a
comon way like via websocket or MQTT.

```
car's OBD2 female plug
 __________________________________________
 \     __  __  __  __  __  __  __  __     /
  \     1   2   3   4   5   6   7   8    /            Hardware
   \      ________________________      / --------->  CANbus chip e.g. MCP2518FD  -->  Linux driver  -->  Network Interface -
    \  __  __  __  __  __  __  __  __  /                                                                                     |
     \  9  10  11  12  13  14  15  16 /        this go-package                                                               |
      \______________________________/    --------------------------------------------------------------------------------   | (Network Interface/Serial/TCP)
                                         |                       human readable                                           |  |
                                         |   forwarders (MQTT/WS)  <--<struct>-->  cancoder  <--<can.Frame>-->  canbus  <-|--
                                         |             ( /\________________________________________| )                    |
                                          --------------------------------------------------------------------------------
```

Supported:
 * Network Interface (e.g. `can0`)
 * Serial (`tty`)
 * TCP

## CAN

### Interfacing via Network Interface

For connecting to the CANbus, a device like a `MCP2515` is required.
see: `https://wiki.seeedstudio.com/2-Channel-CAN-BUS-FD-Shield-for-Raspberry-Pi/`