// +build !windows

package conn

import (
	"fmt"
	"net"
	"os"
	"syscall"
)

//Ny linux
func DialBroadcastUDP(port int) net.PacketConn {
	s, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_UDP)
	if err != nil {
		fmt.Printf("Error: Socket: %#+v\n")
	}
	syscall.SetsockoptInt(s, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
	if err != nil {
		fmt.Printf("Error: SetSockOpt REUSEADDR: %#+v\n")
	}
	syscall.SetsockoptInt(s, syscall.SOL_SOCKET, syscall.SO_BROADCAST, 1)
	if err != nil {
		fmt.Printf("Error: SetSockOpt BROADCAST:  %#+v\n")
	}
	syscall.Bind(s, &syscall.SockaddrInet4{Port: port})
	if err != nil {
		fmt.Printf("Error: Bind:  %#+v\n")
	}

	f := os.NewFile(uintptr(s), "")
	conn, err := net.FilePacketConn(f)
	if err != nil {
		fmt.Printf("Error: FilePacketConn: %#+v\n")
	}
	f.Close()

	return conn
}

/*
//For mac
func DialBroadcastUDP(port int) net.PacketConn {
	s, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_UDP)
	if err != nil { fmt.Printf("Error: Socket: %#+v\n") }
	syscall.SetsockoptInt(s, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
	if err != nil { fmt.Printf("Error: SetSockOpt REUSEADDR: %#+v\n") }
	syscall.SetsockoptInt(s, syscall.SOL_SOCKET, syscall.SO_BROADCAST, 1)
	if err != nil { fmt.Printf("Error: SetSockOpt BROADCAST:  %#+v\n") }
	syscall.SetsockoptInt(s, syscall.SOL_SOCKET, syscall.SO_REUSEPORT, 1)
	if err != nil { fmt.Printf("Error: SetSockOpt REUSEPORT:  %#+v\n") }
	syscall.Bind(s, &syscall.SockaddrInet4{Port: port})
	if err != nil { fmt.Printf("Error: Bind:  %#+v\n") }

	f := os.NewFile(uintptr(s), "")
	conn, err := net.FilePacketConn(f)
	if err != nil { fmt.Printf("Error: FilePacketConn: %#+v\n") }
	f.Close()

	return conn
}






/*
func DialBroadcastUDP(port int) net.PacketConn {
	s, _ := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_UDP)
	syscall.SetsockoptInt(s, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
	syscall.SetsockoptInt(s, syscall.SOL_SOCKET, syscall.SO_BROADCAST, 1)
	syscall.Bind(s, &syscall.SockaddrInet4{Port: port})

	f := os.NewFile(uintptr(s), "")
	conn, _ := net.FilePacketConn(f)
	f.Close()

	return conn
}
*/
