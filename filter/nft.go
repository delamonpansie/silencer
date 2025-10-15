package filter

import (
	"net"
	"strings"
	"time"
)

type nft struct {
	table string
	set   string
}

func NewNFT(table, set string) nft {
	return nft{table, set}
}

func (b nft) Block(ip net.IP, duration time.Duration, name string) {
	command("nft", "add", "element", b.table, b.set,
		"{",
		ip.String(),
		"timeout", duration.String(),
		"comment", strings.ReplaceAll(name, `"`, ``),
		"}",
	)
}
