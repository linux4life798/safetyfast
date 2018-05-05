package main

import "math/rand"

type RandValues []interface{}

func NewRandValues() *RandValues {
	rv := make([]interface{}, 0, 1)
	return (*RandValues)(&rv)
}

func (rv *RandValues) Len() int {
	return len(*rv)
}

func (rv *RandValues) Put(v ...interface{}) *RandValues {
	*rv = append(*rv, v...)
	return rv
}

func (rv *RandValues) GetAll() []interface{} {
	return []interface{}(*rv)
}

func (rv *RandValues) Shuffle() *RandValues {
	rand.Shuffle(len(*rv), func(i, j int) {
		(*rv)[i], (*rv)[j] = (*rv)[j], (*rv)[i]
	})
	return rv
}

func (rv *RandValues) Clear() *RandValues {
	*rv = (*rv)[0:0]
	return rv
}

func (rv *RandValues) AddSparseInt32(count int) *RandValues {
	for i := 0; i < count; i++ {
		*rv = append(*rv, rand.Int31())
	}
	return rv
}

func (rv *RandValues) AddSparseInt64(count int) *RandValues {
	for i := 0; i < count; i++ {
		*rv = append(*rv, rand.Int63())
	}
	return rv
}

func (rv *RandValues) AddUniformInt32(count int, mod int32) *RandValues {
	for i := 0; i < count; i++ {
		*rv = append(*rv, rand.Int31n(mod))
	}
	return rv
}

func (rv *RandValues) AddUniformInt64(count int, mod int64) *RandValues {
	for i := 0; i < count; i++ {
		*rv = append(*rv, rand.Int63n(mod))
	}
	return rv
}

func (rv *RandValues) AddSparseUint32(count int) *RandValues {
	for i := 0; i < count; i++ {
		*rv = append(*rv, rand.Uint32())
	}
	return rv
}

func (rv *RandValues) AddSparseUint64(count int) *RandValues {
	for i := 0; i < count; i++ {
		*rv = append(*rv, rand.Uint64())
	}
	return rv
}

func (rv *RandValues) AddSparseFloat32(count int) *RandValues {
	for i := 0; i < count; i++ {
		*rv = append(*rv, rand.Float32())
	}
	return rv
}

func (rv *RandValues) AddSparseFloat64(count int) *RandValues {
	for i := 0; i < count; i++ {
		*rv = append(*rv, rand.Float64())
	}
	return rv
}
