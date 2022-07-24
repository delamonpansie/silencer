package filter

import (
	"fmt"
	"net"
	"os/exec"

	"github.com/delamonpansie/silencer/logger"
	"go.uber.org/zap"
)

var log = &logger.Log

//go:generate mockgen -destination filter_mock.go -package filter -source filter.go Blocker
type Blocker interface {
	Block(net.IP)
	Unblock(net.IP)
	List() []net.IP
}

type dummy struct{}

func NewDummy() *dummy {
	return &dummy{}
}
func (dummy) Block(ip net.IP)   { fmt.Println("block", ip) }
func (dummy) Unblock(ip net.IP) { fmt.Println("unblock", ip) }
func (dummy) List() []net.IP    { return nil }

func command(command string, args ...string) ([]byte, error) {
	cmd := exec.Command(command, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Warn("command failed", zap.Stringer("command", cmd), zap.String("output", string(output)))
	}
	return output, err
}
