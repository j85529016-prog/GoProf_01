package hw04lrucache

import (
	"math/rand"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCache(t *testing.T) {
	t.Run("empty cache", func(t *testing.T) {
		c := NewCache(10)

		_, ok := c.Get("aaa")
		require.False(t, ok)

		_, ok = c.Get("bbb")
		require.False(t, ok)
	})

	t.Run("simple", func(t *testing.T) {
		c := NewCache(5)

		wasInCache := c.Set("aaa", 100)
		require.False(t, wasInCache)

		wasInCache = c.Set("bbb", 200)
		require.False(t, wasInCache)

		val, ok := c.Get("aaa")
		require.True(t, ok)
		require.Equal(t, 100, val)

		val, ok = c.Get("bbb")
		require.True(t, ok)
		require.Equal(t, 200, val)

		wasInCache = c.Set("aaa", 300)
		require.True(t, wasInCache)

		val, ok = c.Get("aaa")
		require.True(t, ok)
		require.Equal(t, 300, val)

		val, ok = c.Get("ccc")
		require.False(t, ok)
		require.Nil(t, val)
	})

	t.Run("bad capacity", func(t *testing.T) {
		cash := NewCache(0)
		require.Nil(t, cash)
		cash = NewCache(-10)
		require.Nil(t, cash)
		cash = NewCache(1)
		require.NotNil(t, cash)
	})

	t.Run("complex", func(t *testing.T) {
		cash := NewCache(4)
		cash.Set("k1", 1)          // "1"
		val, ok := cash.Get("key") // "1"
		require.False(t, ok)
		require.Nil(t, val)
		ok = cash.Set("k2", 2) // "2 ↔ 1"
		require.False(t, ok)
		cash.Set("k3", 3)        // "3 ↔ 2 ↔ 1"
		cash.Set("k4", 4)        // "4 ↔ 3 ↔ 2 ↔ 1"
		val, ok = cash.Get("k1") // "1 ↔ 4 ↔ 3 ↔ 2 "
		require.True(t, ok)
		require.Equal(t, 1, val)
		cash.Set("k5", 5)        // "5 ↔ 1 ↔ 4 ↔ 3"
		val, ok = cash.Get("k2") // "5 ↔ 1 ↔ 4 ↔ 3"
		require.False(t, ok)
		require.Nil(t, val)
		ok = cash.Set("k1", 9999) // "9999 ↔ 5 ↔ 4 ↔ 3"
		require.True(t, ok)
		cash.Set("k6", 6)        // "6 ↔ 9999 ↔ 5 ↔ 4"
		val, ok = cash.Get("k3") // "6 ↔ 9999 ↔ 5 ↔ 4"
		require.False(t, ok)
		require.Nil(t, val)
		cash.Clear() // ""
		_, ok = cash.Get("k6")
		require.False(t, ok)
		_, ok = cash.Get("k1")
		require.False(t, ok)
		_, ok = cash.Get("k5")
		require.False(t, ok)
		_, ok = cash.Get("k4")
		require.False(t, ok)
	})
}

func TestCacheMultithreading(t *testing.T) {
	c := NewCache(10)
	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < 1_000_000; i++ {
			c.Set(Key(strconv.Itoa(i)), i)
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 1_000_000; i++ {
			c.Get(Key(strconv.Itoa(rand.Intn(1_000_000))))
		}
	}()

	wg.Wait()

	require.NotNil(t, c)
}
