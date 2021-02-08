package filter

import (
	"log"
	"net"
	"os/exec"
	"strings"
)

type ipset struct {
	set string
}

func NewIPset(set string) *ipset {
	return &ipset{set}
}

func (b ipset) Block(ip net.IP) {
	cmd := exec.Command("ipset", "add", b.set, ip.String())
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("command %q failed with %q", cmd, string(output))
	}
}

func (b ipset) Unblock(ip net.IP) {
	cmd := exec.Command("ipset", "del", b.set, ip.String())
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("command %q failed with %q", cmd, string(output))
	}
}

func parseIpsetList(buf []byte) (list []net.IP) {
	for _, line := range strings.Split(string(buf), "\n") {
		fields := strings.Fields(line)
		if len(fields) != 3 {
			continue
		}
		if fields[0] != "add" {
			continue
		}
		ip := net.ParseIP(fields[2]).To4()
		if ip == nil {
			log.Printf("invalid IPv4 address: %q", fields[2])
			continue
		}
		list = append(list, ip)
	}
	return
}

func (b ipset) List() []net.IP {
	cmd := exec.Command("ipset", "list", b.set, "-format", "save")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("command %q failed with %q", cmd, string(output))
		return nil
	}

	return parseIpsetList(output)
}
