package filter

import (
	"net"
	"os/exec"
	"strings"

	"go.uber.org/zap"
)

type ipset struct {
	set string
}

func NewIPset(set string) *ipset {
	cmd := exec.Command("ipset", "create", set, "hash:ip")
	output, err := cmd.CombinedOutput()
	if err != nil && !strings.Contains(string(output), "set with the same name already exists") {
		log.Warn("command failed", zap.Stringer("command", cmd), zap.String("output", string(output)))
	}
	return &ipset{set}
}

func (b ipset) Block(ip net.IP) {
	command("ipset", "add", b.set, ip.String())
}

func (b ipset) Unblock(ip net.IP) {
	command("ipset", "del", b.set, ip.String())
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
			log.Warn("invalid IPv4 address", zap.Any("ip", fields[2]))
			continue
		}
		list = append(list, ip)
	}
	return
}

func (b ipset) List() []net.IP {
	output, err := command("ipset", "list", b.set, "-output", "save")
	if err != nil {
		return nil
	}

	return parseIpsetList(output)
}
