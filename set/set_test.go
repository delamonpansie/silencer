package set

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	a = net.ParseIP("1.1.1.1").To4()
	b = net.ParseIP("2.2.2.2").To4()
	c = net.ParseIP("3.3.3.3").To4()
	d = net.ParseIP("4.4.4.4").To4()

	e = net.IP{1, 2, 3}
)

func ipv4(ip net.IP) (ret [4]byte) {
	copy(ret[:], ip)
	return
}

func Test_pset_Push_pushes_to_tail(t *testing.T) {
	h := pset{set: make(map[[4]byte]int)}
	h.Push(item{ip: ipv4(a)})
	assert.Equal(t, ipv4(a), h.heap[0].ip)
	assert.Equal(t, 0, h.set[ipv4(a)])

	h.Push(item{ip: ipv4(b)})
	assert.Equal(t, ipv4(b), h.heap[1].ip)
	assert.Equal(t, 1, h.set[ipv4(b)])
	fmt.Println(e)
}

func Test_pset_Len(t *testing.T) {
	h := pset{set: make(map[[4]byte]int)}
	assert.Equal(t, 0, h.Len())
	h.Push(item{})
	assert.Equal(t, 1, h.Len())
}

func Test_pset_Swap_swaps_elements(t *testing.T) {
	h := pset{set: make(map[[4]byte]int)}
	h.Push(item{ip: ipv4(a)})
	h.Push(item{ip: ipv4(b)})

	h.Swap(0, 1)
	assert.Equal(t, ipv4(a), h.heap[1].ip)
	assert.Equal(t, ipv4(b), h.heap[0].ip)
	assert.Equal(t, 1, h.set[ipv4(a)])
	assert.Equal(t, 0, h.set[ipv4(b)])

	// second swap reverses change
	h.Swap(1, 0)
	assert.Equal(t, ipv4(a), h.heap[0].ip)
	assert.Equal(t, ipv4(b), h.heap[1].ip)
	assert.Equal(t, 0, h.set[ipv4(a)])
	assert.Equal(t, 1, h.set[ipv4(b)])
}

func Test_pset_Pop_pops_last_element(t *testing.T) {
	h := pset{set: make(map[[4]byte]int)}
	h.Push(item{ip: ipv4(a)})
	h.Push(item{ip: ipv4(b)})
	el := h.Pop().(item)
	assert.Equal(t, ipv4(b), el.ip)
	assert.Equal(t, ipv4(a), h.heap[0].ip)
	assert.Equal(t, 0, h.set[ipv4(a)])
	_, exist := h.set[ipv4(b)]
	assert.False(t, exist)
}

func Test_Set_Insert(t *testing.T) {
	t.Run("heap[0] contains earliest deadline", func(t *testing.T) {
		s := NewSet()
		new := s.Insert(a, time.Millisecond*10)
		assert.True(t, new)
		assert.Equal(t, ipv4(a), s.inner.heap[0].ip)

		new = s.Insert(b, time.Millisecond*5)
		assert.True(t, new)
		assert.Equal(t, ipv4(b), s.inner.heap[0].ip)

		new = s.Insert(c, time.Millisecond*2)
		assert.True(t, new)
		assert.Equal(t, ipv4(c), s.inner.heap[0].ip)

		new = s.Insert(d, time.Millisecond*20)
		assert.True(t, new)
		assert.Equal(t, ipv4(c), s.inner.heap[0].ip)
	})

	t.Run("updates deadline", func(t *testing.T) {
		s := NewSet()
		new := s.Insert(a, time.Millisecond*5)
		assert.True(t, new)
		assert.Equal(t, ipv4(a), s.inner.heap[0].ip)

		new = s.Insert(b, time.Millisecond*20)
		assert.True(t, new)
		assert.Equal(t, ipv4(a), s.inner.heap[0].ip)

		// after update a, b should become first
		new = s.Insert(a, time.Minute)
		assert.False(t, new)
		assert.Equal(t, ipv4(b), s.inner.heap[0].ip)
	})
}

func Test_Set_Expire(t *testing.T) {
	s := NewSet()
	new := s.Insert(a, time.Millisecond*1)
	assert.True(t, new)
	assert.Equal(t, ipv4(a), s.inner.heap[0].ip)

	new = s.Insert(b, time.Millisecond*2)
	assert.True(t, new)
	assert.Equal(t, ipv4(a), s.inner.heap[0].ip)

	new = s.Insert(c, time.Minute)
	assert.True(t, new)
	assert.Equal(t, ipv4(a), s.inner.heap[0].ip)

	time.Sleep(time.Millisecond * 100)

}

func Test_Set_Deadline(t *testing.T) {
	s := NewSet()

	assert.Equal(t, time.Time{}, s.Deadline())

	s.Insert(a, time.Second)
	assert.Less(t, time.Now().Sub(s.Deadline()), time.Millisecond)
}
