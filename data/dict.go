package data

// Dict is a concurrency-safe map with string keys. It is a shortcut for Map[string, V].
type Dict[V any] = Map[string, V]

// NewDict returns an initialized Dict.
func NewDict[V any]() *Dict[V] {
	return NewMap[string, V]()
}
