package main


import (
	
	"fmt"
	"net"
	
)

func Recive() {
  pc,err := net.ListenPacket("udp4", ":30000")
  if err != nil {
    panic(err)
  }
  defer pc.Close()

  buf := make([]byte, 1024)
  for {
	  n,addr,err := pc.ReadFrom(buf)
	  if err != nil {
	    panic(err)
	  }
  
  	fmt.Printf("%s sent this: %s\n", addr, buf[:n])
  }
}
