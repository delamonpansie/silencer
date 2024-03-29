package set

import (
	"container/heap"
	"fmt"
	"log"
	"net"
	"time"
)

type ip [4]byte

func (ip ip) String() string {
	return fmt.Sprintf("%d.%d.%d.%d", ip[0], ip[1], ip[2], ip[3])
}

type item struct {
	ip       ip
	deadline time.Time
}

// pset implements priority set
type pset struct {
	heap []item
	set  map[ip]int
}

var _ heap.Interface = &pset{}

func (b pset) Len() int {
	if len(b.heap) != len(b.set) {
		log.Fatalf("set.Len: invariant violation, len(heap) = %d, but len(set) = %d",
			len(b.heap), len(b.set))
	}
	return len(b.heap)
}

func (b pset) Less(i, j int) bool {
	return b.heap[i].deadline.Before(b.heap[j].deadline)
}

func (b pset) Swap(i, j int) {
	bi, bj := b.heap[i], b.heap[j]
	b.heap[i], b.heap[j] = bj, bi
	b.set[bi.ip] = j
	b.set[bj.ip] = i
}

func (b *pset) Push(i interface{}) {
	x := i.(item)
	b.heap = append(b.heap, x)
	b.set[x.ip] = len(b.heap) - 1
}

func (b *pset) Pop() interface{} {
	n := len(b.heap)
	x := b.heap[n-1]
	b.heap = b.heap[:n-1]
	delete(b.set, x.ip)
	return x
}

// Set implements set with memberership expiration.
type Set struct {
	// embedding pset without name will leak Swap/Push/Pop methods
	inner pset
}

func NewSet() Set {
	return Set{pset{set: make(map[ip]int)}}
}

func ip4(ip net.IP) (addr ip) {
	ip4 := ip.To4()
	if len(ip4) != 4 {
		log.Fatalf("invalid IPv4: %q", ip)
	}
	copy(addr[:], ip4)
	return
}

// Insert will insert element into set or update duration if element already exists
func (s *Set) Insert(netIP net.IP, duration time.Duration) bool {
	ip := ip4(netIP)

	b := &s.inner
	deadline := time.Now().Add(duration)
	if i, exists := b.set[ip]; exists {
		old := &b.heap[i]
		if old.ip != ip {
			log.Fatalf("Set.Insert: invariant violation, heap[i].ip = %s, but should be %s",
				old.ip, ip)
		}
		if old.deadline.Before(deadline) {
			old.deadline = deadline
			heap.Fix(b, i)
		}
		return false
	} else {
		heap.Push(b, item{ip, deadline})
		return true
	}
}

// Expire must be called periodically to purge old entries
// returns expired entries
func (s *Set) Expire() (expired []net.IP) {
	b := &s.inner
	now := time.Now()
	for len(b.heap) > 0 && b.heap[0].deadline.Before(now) {
		ip := heap.Pop(b).(item).ip
		expired = append(expired, ip[:])
	}
	return
}

// Deadline retunrn first deadline
func (s *Set) Deadline() time.Time {
	b := &s.inner
	if len(b.heap) > 0 {
		return b.heap[0].deadline
	}
	return time.Time{}
}
