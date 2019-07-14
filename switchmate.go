package main

import (
  "os"
  "log"
  "fmt"
  "github.com/paypal/gatt"
)

var uartServiceId = gatt.MustParseUUID("a22bd383-ebdd-49ac-b2e7-40eb55f5d0ab")
var uartServiceTXCharId = gatt.MustParseUUID("a22b0070-ebdd-49ac-b2e7-40eb55f5d0ab")
var uartServiceRXCharId = gatt.MustParseUUID("a22b0090-ebdd-49ac-b2e7-40eb55f5d0ab")

func onStateChanged(d gatt.Device, s gatt.State) {
  fmt.Println("State:", s)
  switch s {
  case gatt.StatePoweredOn:
    fmt.Println("scanning...")
    d.Scan([]gatt.UUID{}, false)
    return
  default:
    d.StopScanning()
  }
}

func onPeriphDiscovered(p gatt.Peripheral, a *gatt.Advertisement, rssi int) {
  if (p.ID() == deviceId) {
    p.Device().StopScanning()
    p.Device().Connect(p);
  } else if (deviceId == "discover") {
    fmt.Printf("Preipheral Discovered: %s \n", p.ID())
  }
}

func onPeriphConnected(p gatt.Peripheral, err error) {
  fmt.Printf("Peripheral connected\n")

  services, err := p.DiscoverServices(nil)
  if err != nil {
    log.Printf("Failed to discover services, err: %s\n", err)
    return
  }

  for _, service := range services {
    if (service.UUID().Equal(uartServiceId)) {
      fmt.Printf("Service Found %s\n", service.Name())
      cs, _ := p.DiscoverCharacteristics(nil, service)
      for _, c := range cs {
        if (c.UUID().Equal(uartServiceTXCharId)) {
          fmt.Println("TX Characteristic Found")
          p.DiscoverDescriptors(nil, c)
          p.SetNotifyValue(c, func(c *gatt.Characteristic, b []byte, e error) {
            fmt.Printf("Got back %s\n", string(b))
          })
        }
        if (c.UUID().Equal(uartServiceRXCharId)) {
          fmt.Println("RX Characteristic Found")
          p.WriteCharacteristic(c, state, true)
          fmt.Printf("Wrote %s\n", string(state))
        }
      }
    }
  }
}

func onPeriphDisconnected(p gatt.Peripheral, err error) {
  fmt.Println("Disconnected")
  done <- true
}

var done = make(chan bool)

var deviceId string
var state []byte

func main() {
  deviceId = os.Args[1]

  if deviceId != "discover" {
    flag := os.Args[2]

    if flag == "on" {
      state = []byte{0x01}
    } else {
      state = []byte{0x00}
    }
  }

  var DefaultClientOptions = []gatt.Option{
    gatt.LnxMaxConnections(1),
    gatt.LnxDeviceID(-1, false),
  }

  d, err := gatt.NewDevice(DefaultClientOptions...)
  if err != nil {
    log.Fatalf("Failed to open device, err: %s\n", err)
  }

  d.Handle(
    gatt.PeripheralDiscovered(onPeriphDiscovered),
    gatt.PeripheralConnected(onPeriphConnected),
    gatt.PeripheralDisconnected(onPeriphDisconnected),
  )

  d.Init(onStateChanged)
  <-done
  log.Println("Done")
}
