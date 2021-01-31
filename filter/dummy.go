package filter

import (
	"fmt"
)

type dummy struct{}

func NewDummy() *dummy {
	return &dummy{}
}
func (dummy) Block(ip string)   { fmt.Println("block", ip) }
func (dummy) Unblock(ip string) { fmt.Println("unblock", ip) }
func (dummy) List() []string    { return nil }
