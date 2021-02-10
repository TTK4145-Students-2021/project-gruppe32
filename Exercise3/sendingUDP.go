package main

import (
  "net"
  "fmt"
  
)

func main() {
  pc, err := net.ListenPacket("udp4", ":20014")
  sender, err := net.ListenPacket("udp4", "")
  if err != nil {
    panic(err)
  }
  defer pc.Close()

  serverAddr,err := net.ResolveUDPAddr("udp4", "10.100.23.147:20014")
  if err != nil {
    panic(err)
  }

  _,err = sender.WriteTo([]byte("data to transmit"), serverAddr)
  if err != nil {
    panic(err)
  }


  // listen
  buf := make([]byte, 1024)
  
  n, _,err := pc.ReadFrom(buf)
  if err != nil {
    panic(err)
  }
  fmt.Printf("The server replied this: %s\n", buf[:n])
}


