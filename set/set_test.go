package set

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_pset_Push_pushes_to_tail(t *testing.T) {
	h := pset{set: make(map[string]int)}
	h.Push(item{ip: "a"})
	assert.Equal(t, "a", h.heap[0].ip)
	assert.Equal(t, 0, h.set["a"])

	h.Push(item{ip: "b"})
	assert.Equal(t, "b", h.heap[1].ip)
	assert.Equal(t, 1, h.set["b"])
}

func Test_pset_Len(t *testing.T) {
	h := pset{set: make(map[string]int)}
	assert.Equal(t, 0, h.Len())
	h.Push(item{})
	assert.Equal(t, 1, h.Len())
}

func Test_pset_Swap_swaps_elements(t *testing.T) {
	h := pset{set: make(map[string]int)}
	h.Push(item{ip: "a"})
	h.Push(item{ip: "b"})

	h.Swap(0, 1)
	assert.Equal(t, "a", h.heap[1].ip)
	assert.Equal(t, "b", h.heap[0].ip)
	assert.Equal(t, 1, h.set["a"])
	assert.Equal(t, 0, h.set["b"])

	// second swap reverses change
	h.Swap(1, 0)
	assert.Equal(t, "a", h.heap[0].ip)
	assert.Equal(t, "b", h.heap[1].ip)
	assert.Equal(t, 0, h.set["a"])
	assert.Equal(t, 1, h.set["b"])
}

func Test_pset_Pop_pops_last_element(t *testing.T) {
	h := pset{set: make(map[string]int)}
	h.Push(item{ip: "a"})
	h.Push(item{ip: "b"})
	el := h.Pop().(item)
	assert.Equal(t, "b", el.ip)
	assert.Equal(t, "a", h.heap[0].ip)
	assert.Equal(t, 0, h.set["a"])
	_, exist := h.set["b"]
	assert.False(t, exist)
}

func Test_Set_Insert(t *testing.T) {
	t.Run("heap[0] contains earliest deadline", func(t *testing.T) {
		s := NewSet()
		new := s.Insert("a", time.Millisecond*10)
		assert.True(t, new)
		assert.Equal(t, "a", s.inner.heap[0].ip)

		new = s.Insert("b", time.Millisecond*5)
		assert.True(t, new)
		assert.Equal(t, "b", s.inner.heap[0].ip)

		new = s.Insert("c", time.Millisecond*2)
		assert.True(t, new)
		assert.Equal(t, "c", s.inner.heap[0].ip)

		new = s.Insert("d", time.Millisecond*20)
		assert.True(t, new)
		assert.Equal(t, "c", s.inner.heap[0].ip)
	})

	t.Run("updates deadline", func(t *testing.T) {
		s := NewSet()
		new := s.Insert("a", time.Millisecond*5)
		assert.True(t, new)
		assert.Equal(t, "a", s.inner.heap[0].ip)

		new = s.Insert("b", time.Millisecond*20)
		assert.True(t, new)
		assert.Equal(t, "a", s.inner.heap[0].ip)

		// after update "a", "b" should become first
		new = s.Insert("a", time.Minute)
		assert.False(t, new)
		assert.Equal(t, "b", s.inner.heap[0].ip)
	})
}

func Test_Set_Expire(t *testing.T) {
	s := NewSet()
	new := s.Insert("a", time.Millisecond*1)
	assert.True(t, new)
	assert.Equal(t, "a", s.inner.heap[0].ip)

	new = s.Insert("b", time.Millisecond*2)
	assert.True(t, new)
	assert.Equal(t, "a", s.inner.heap[0].ip)

	new = s.Insert("c", time.Minute)
	assert.True(t, new)
	assert.Equal(t, "a", s.inner.heap[0].ip)

	time.Sleep(time.Millisecond * 100)

}
