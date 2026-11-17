package safe

import "sync"

type List[T any] struct {
	Mx   sync.Mutex
	List []T
}

func (sList *List[T]) Dequeue() T {
	sList.Mx.Lock()
	var value = sList.List[0]
	sList.List = sList.List[1:]
	sList.Mx.Unlock()

	return value
}
func (sList *List[T]) Enqueue(value T) {
	sList.Mx.Lock()
	sList.List = append(sList.List, value)
	sList.Mx.Unlock()
}
