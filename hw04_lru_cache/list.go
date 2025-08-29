package hw04lrucache

import (
	"fmt"
	"strings"
)

type List interface {
	Len() int
	Front() *ListItem
	Back() *ListItem
	PushFront(v interface{}) *ListItem
	PushBack(v interface{}) *ListItem
	Remove(i *ListItem)
	MoveToFront(i *ListItem)
}

type ListItem struct {
	Value interface{}
	Next  *ListItem
	Prev  *ListItem
}

type linkedList struct {
	Size      int
	NodeFront *ListItem
	NodeBack  *ListItem
}

func NewList() List {
	return &linkedList{}
}

func (l *linkedList) Len() int {
	return l.Size
}

func (l *linkedList) Front() *ListItem {
	return l.NodeFront
}

func (l *linkedList) Back() *ListItem {
	return l.NodeBack
}

func (l *linkedList) PushFront(v interface{}) *ListItem {
	// Создаем listItem со значением, но без привязки
	listItem := ListItem{Value: v, Next: nil, Prev: nil}

	// Если размер связанного списка = 0
	if l.Size == 0 {
		l.NodeFront = &listItem
		l.NodeBack = &listItem
	} else {
		listItem.Next = l.NodeFront
		l.NodeFront.Prev = &listItem
		l.NodeFront = &listItem
	}

	l.Size++

	return &listItem
}

func (l *linkedList) PushBack(v interface{}) *ListItem {
	// Создаем listItem со значением, но без привязки
	listItem := ListItem{Value: v, Next: nil, Prev: nil}

	// Если размер связанного списка = 0
	if l.Size == 0 {
		l.NodeFront = &listItem
		l.NodeBack = &listItem
		l.Size++
		return &listItem
	}

	// К listItem.Prev привязываем первый элемент списка linkedList.NodeBack
	listItem.Prev = l.NodeBack
	l.NodeBack.Next = &listItem
	l.NodeBack = &listItem
	l.Size++
	return &listItem
}

func (l *linkedList) Remove(i *ListItem) {
	// Если пустой параметр или список пустой - ничего не делаем
	if i == nil || l.Size == 0 {
		return
	}

	// Обновляем связи соседних элементов
	if i.Prev != nil {
		i.Prev.Next = i.Next
	}
	if i.Next != nil {
		i.Next.Prev = i.Prev
	}

	// Обновляем границы списка
	if i == l.NodeFront {
		l.NodeFront = i.Next
	}
	if i == l.NodeBack {
		l.NodeBack = i.Prev
	}

	// Очищаем связи удаляемого элемента
	i.Next = nil
	i.Prev = nil

	// Уменьшаем длину списка в отложенном вызове
	l.Size--
}

func (l *linkedList) MoveToFront(i *ListItem) {
	// Если пустой параметр, или элемент уже занимает front-позицию
	// или список длиной <= 1 - ничего не делаем
	if i == nil || i == l.NodeFront || l.Size <= 1 {
		return
	}

	// Сохраняем оригинальные указатели
	prev := i.Prev
	next := i.Next

	// Обновляем соседние элементы
	if prev != nil {
		prev.Next = next
	}
	if next != nil {
		next.Prev = prev
	}

	// Обновляем NodeBack, если перемещаемый элемент был последним
	if i == l.NodeBack {
		l.NodeBack = prev
	}

	// Вставляем в начало
	i.Prev = nil
	i.Next = l.NodeFront
	l.NodeFront.Prev = i
	l.NodeFront = i
}

func (l *linkedList) String() string {
	if l.NodeFront == nil {
		return "nil"
	}

	var parts []string
	current := l.NodeFront

	for current != nil {
		parts = append(parts, fmt.Sprintf("%v", current.Value))
		current = current.Next
	}

	return strings.Join(parts, " ↔ ")
}
