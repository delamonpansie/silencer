package filter

import (
	"fmt"
	"net"
	"os/exec"
	"time"

	"github.com/delamonpansie/silencer/logger"
	"go.uber.org/zap"
)

var log = &logger.Log

//go:generate mockgen -destination filter_mock.go -package filter -source filter.go Blocker
type Blocker interface {
	// Block the ip for given duration. Blocker implementation is responsible for removing the block
	// after duration has passed.
	Block(net.IP, time.Duration)
}

type dummy struct{}

func NewDummy() *dummy {
	return &dummy{}
}
func (dummy) Block(ip net.IP, duration time.Duration) {
	fmt.Println("block", ip, duration)
}

func command(command string, args ...string) ([]byte, error) {
	cmd := exec.Command(command, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Warn("command failed", zap.Stringer("command", cmd), zap.String("output", string(output)))
	}
	return output, err
}
