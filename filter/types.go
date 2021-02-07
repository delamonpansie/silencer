package filter

import (
	"net"
)

//go:generate mockgen -destination types_mock.go -package filter -source types.go Blocker

type Blocker interface {
	Block(net.IP)
	Unblock(net.IP)
	List() []net.IP
}
