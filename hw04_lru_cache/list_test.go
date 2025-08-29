package hw04lrucache

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestList(t *testing.T) {
	t.Run("empty list", func(t *testing.T) {
		l := NewList()

		require.Equal(t, 0, l.Len())
		require.Nil(t, l.Front())
		require.Nil(t, l.Back())
	})

	t.Run("complex", func(t *testing.T) {
		l := NewList()

		l.PushFront(10) // [10]
		l.PushBack(20)  // [10, 20]
		l.PushBack(30)  // [10, 20, 30]
		require.Equal(t, 3, l.Len())

		middle := l.Front().Next // 20
		l.Remove(middle)         // [10, 30]
		require.Equal(t, 2, l.Len())

		for i, v := range [...]int{40, 50, 60, 70, 80} {
			if i%2 == 0 {
				l.PushFront(v)
			} else {
				l.PushBack(v)
			}
		} // [80, 60, 40, 10, 30, 50, 70]

		require.Equal(t, 7, l.Len())
		require.Equal(t, 80, l.Front().Value)
		require.Equal(t, 70, l.Back().Value)

		l.MoveToFront(l.Front()) // [80, 60, 40, 10, 30, 50, 70]
		l.MoveToFront(l.Back())  // [70, 80, 60, 40, 10, 30, 50]

		elems := make([]int, 0, l.Len())
		for i := l.Front(); i != nil; i = i.Next {
			elems = append(elems, i.Value.(int))
		}
		require.Equal(t, []int{70, 80, 60, 40, 10, 30, 50}, elems)
	})

	t.Run("relations between elements", func(t *testing.T) {
		list := NewList()
		elementForRemove := list.PushBack("A")       // "A"
		list.PushBack("B")                           // "A ↔ B"
		elementForMoveToFront := list.PushFront("C") // "C ↔ A ↔ B"
		list.PushBack(123)                           // "C ↔ A ↔ B ↔ 123"
		list.PushFront(1)                            // "1 ↔ C ↔ A ↔ B ↔ 123"
		list.Remove(elementForRemove)                // "1 ↔ C ↔ B ↔ 123"
		list.MoveToFront(elementForMoveToFront)      // "C ↔ 1 ↔ B ↔ 123"
		list.MoveToFront(list.Back())                // "123 ↔ C ↔ 1 ↔ B"
		require.Equal(t, list.Back().Value, "B")
		require.Equal(t, list.Front().Value, 123)
		require.Equal(t, list.Len(), 4)
		require.Equal(t, fmt.Sprint(list), "123 ↔ C ↔ 1 ↔ B")
	})
}
