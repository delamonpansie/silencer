package filter

import (
	"fmt"
	"net"
)

type dummy struct{}

func NewDummy() *dummy {
	return &dummy{}
}
func (dummy) Block(ip net.IP)   { fmt.Println("block", ip) }
func (dummy) Unblock(ip net.IP) { fmt.Println("unblock", ip) }
func (dummy) List() []net.IP    { return nil }
