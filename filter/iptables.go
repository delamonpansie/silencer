package filter

import (
	"log"
	"os/exec"
	"strings"
)

type iptables struct {
	chain string
}

func NewIPtables(chain string) *iptables {
	return &iptables{chain}
}

func (b iptables) Block(ip string) {
	cmd := exec.Command("iptables", "-I", b.chain, "--src", ip, "-j", "DROP")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("command %q failed with %q", cmd, string(output))
	}
}

func (b iptables) Unblock(ip string) {
	cmd := exec.Command("iptables", "-D", b.chain, "--src", ip, "-j", "DROP")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("command %q failed with %q", cmd, string(output))
	}
}

func parseIptablesList(buf []byte) (ip []string) {
	for _, line := range strings.Split(string(buf), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		if fields[0] != "DROP" {
			continue
		}
		ip = append(ip, fields[3])
	}
	return
}

func (b iptables) List() []string {
	cmd := exec.Command("iptables", "-nL", b.chain)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("command %q failed with %q", cmd, string(output))
		return nil
	}

	return parseIptablesList(output)
}
