package filter

import (
	"log"
	"net"
	"strings"
)

type iptables struct {
	chain string
}

func NewIPtables(chain string) *iptables {
	return &iptables{chain}
}

func (b iptables) Block(ip net.IP) {
	command("iptables", "-I", b.chain, "--src", ip.String(), "-j", "DROP")
}

func (b iptables) Unblock(ip net.IP) {
	command("iptables", "-D", b.chain, "--src", ip.String(), "-j", "DROP")
}

func parseIptablesList(buf []byte) (list []net.IP) {
	for _, line := range strings.Split(string(buf), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		if fields[0] != "DROP" {
			continue
		}
		ip := net.ParseIP(fields[3]).To4()
		if ip == nil {
			log.Printf("invalid IPv4 address: %q", fields[3])
			continue
		}
		list = append(list, ip)
	}
	return
}

func (b iptables) List() []net.IP {
	output, err := command("iptables", "-nL", b.chain)
	if err != nil {
		return nil
	}
	return parseIptablesList(output)
}
