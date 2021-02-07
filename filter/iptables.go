package filter

import (
	"log"
	"net"
	"os/exec"
	"strings"
)

type iptables struct {
	chain string
}

func NewIPtables(chain string) *iptables {
	return &iptables{chain}
}

func (b iptables) Block(ip net.IP) {
	cmd := exec.Command("iptables", "-I", b.chain, "--src", ip.String(), "-j", "DROP")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("command %q failed with %q", cmd, string(output))
	}
}

func (b iptables) Unblock(ip net.IP) {
	cmd := exec.Command("iptables", "-D", b.chain, "--src", ip.String(), "-j", "DROP")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("command %q failed with %q", cmd, string(output))
	}
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
	cmd := exec.Command("iptables", "-nL", b.chain)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("command %q failed with %q", cmd, string(output))
		return nil
	}

	return parseIptablesList(output)
}
