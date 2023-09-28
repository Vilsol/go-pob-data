package loader

type CalculateFunc[Key comparable, Value any] func(key Key) Value

type ComputationCache[Key comparable, Value any] struct {
	Calculate CalculateFunc[Key, Value]
	Data      map[Key]Value
}

func (c *ComputationCache[Key, Value]) Get(key Key) Value {
	if value, ok := c.Data[key]; ok {
		return value
	}

	value := c.Calculate(key)
	c.Data[key] = value
	return value
}

func NewComputationCache[Key comparable, Value any](f CalculateFunc[Key, Value]) *ComputationCache[Key, Value] {
	return &ComputationCache[Key, Value]{
		Calculate: f,
		Data:      make(map[Key]Value),
	}
}
