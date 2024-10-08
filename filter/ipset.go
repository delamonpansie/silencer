package filter

import (
	"net"
	"strconv"
	"time"
)

type ipset struct {
	set string
}

func NewIPset(set string) ipset {
	command("ipset", "-exist", "create", set, "hash:ip", "timeout", "604800") // 7d
	return ipset{set}
}

func (b ipset) Block(ip net.IP, duration time.Duration) {
	timeout := strconv.Itoa(int(duration.Seconds()))
	command("ipset", "-exist", "add", b.set, ip.String(), "timeout", timeout)
}
