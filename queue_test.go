package funnelqueue

import (
	"runtime"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmptyByDefault(t *testing.T) {
	q := New()
	assert.Empty(t, q.IsEmpty())
}

func TestSimplePushPop(t *testing.T) {
	q := New()
	q.Push("test")
	require.NotEmpty(t, q.IsEmpty())
	v := q.Pop()
	require.NotNil(t, v)
	rv, ok := v.(string)
	require.True(t, ok)
	require.NotNil(t, rv)
	require.EqualValues(t, "test", rv)
}

func TestMultipleGoroutinesWaitFinish(t *testing.T) {
	n := 1000
	wg := sync.WaitGroup{}
	wg.Add(n)
	q := New()

	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			for j := 0; j < n; j++ {
				q.Push(i*n + j)
			}
		}(i)
	}
	wg.Wait()

	m := make(map[int]int, 0)
	for {
		v := q.Pop()
		if v == nil {
			break
		}
		rv, ok := v.(int)
		require.True(t, ok)
		require.NotNil(t, rv)
		m[rv] = 0
	}
	require.Equal(t, n*n, len(m))
}

func TestMultipleGoroutines(t *testing.T) {
	n := 1000
	wg := sync.WaitGroup{}
	wg.Add(n + 1)
	q := New()

	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			for j := 0; j < n; j++ {
				q.Push(i*n + j)
			}
		}(i)
	}

	go func() {
		runtime.Gosched()
		defer wg.Done()
		m := make(map[int]int, 0)
		for {
			v := q.Pop()
			if v == nil {
				break
			}
			rv, ok := v.(int)
			require.True(t, ok)
			require.NotNil(t, rv)
			m[rv] = 0
		}
		require.Equal(t, n*n, len(m))

	}()

	wg.Wait()
}
