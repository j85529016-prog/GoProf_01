package hw04lrucache

import "sync"

type Key string

type Cache interface {
	Set(key Key, value interface{}) bool
	Get(key Key) (interface{}, bool)
	Clear()
}

type lruCache struct {
	capacity      int
	queue         List
	items         map[Key]*ListItem
	itemsReversed map[*ListItem]Key
	mutex         sync.Mutex
}

func (lc *lruCache) Set(key Key, value interface{}) bool {
	// Блокируем мютекс для потокобезопасности и разблокируем в отложенном вызове
	lc.mutex.Lock()
	defer lc.mutex.Unlock()

	// если элемент присутствует в словаре, то обновить его значение
	// и переместить элемент в начало очереди
	if item, exists := lc.items[key]; exists {
		item.Value = value
		lc.queue.MoveToFront(item)
		return true
	}

	// Если размер очереди станет больше  емкости
	// Удаляем последний элемент очереди и его значение из словаря
	if lc.queue.Len()+1 > lc.capacity {
		itemToDelete := lc.queue.Back()
		itemKeyToDelete := lc.itemsReversed[itemToDelete]
		delete(lc.items, itemKeyToDelete)
		delete(lc.itemsReversed, itemToDelete)
		lc.queue.Remove(itemToDelete)
	}

	// Вставляем элемент в очередь и добавляем в элементы кэша
	newListItem := lc.queue.PushFront(value)
	lc.items[key] = newListItem
	lc.itemsReversed[newListItem] = key

	return false
}

func (lc *lruCache) Get(key Key) (interface{}, bool) {
	// Блокируем мютекс для потокобезопасности и разблокируем в отложенном вызове
	lc.mutex.Lock()
	defer lc.mutex.Unlock()

	// Если ключ сущаствует - возвращаем значение элемента и false
	if item, keyExist := lc.items[key]; keyExist {
		lc.queue.MoveToFront(item)
		return item.Value, true
	}

	return nil, false
}

func (lc *lruCache) Clear() {
	// Блокируем мютекс для потокобезопасности и разблокируем в отложенном вызове
	lc.mutex.Lock()
	defer lc.mutex.Unlock()

	lc.queue = NewList()
	lc.items = make(map[Key]*ListItem, lc.capacity)
	lc.itemsReversed = make(map[*ListItem]Key, lc.capacity)
}

func NewCache(capacity int) Cache {
	if capacity <= 0 {
		return nil
	}
	return &lruCache{
		capacity:      capacity,
		queue:         NewList(),
		items:         make(map[Key]*ListItem, capacity),
		itemsReversed: make(map[*ListItem]Key, capacity),
	}
}
