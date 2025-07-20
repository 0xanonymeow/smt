package pool

import (
	"math/big"
	"sync"
)

// BigIntPool provides a pool of reusable big.Int instances to reduce allocations
type BigIntPool struct {
	pool sync.Pool
}

// NewBigIntPool creates a new BigIntPool
func NewBigIntPool() *BigIntPool {
	return &BigIntPool{
		pool: sync.Pool{
			New: func() interface{} {
				return new(big.Int)
			},
		},
	}
}

// Get retrieves a big.Int from the pool
func (p *BigIntPool) Get() *big.Int {
	return p.pool.Get().(*big.Int)
}

// Put returns a big.Int to the pool after resetting it
func (p *BigIntPool) Put(x *big.Int) {
	if x != nil {
		x.SetInt64(0) // Reset to zero
		p.pool.Put(x)
	}
}

// GetCopy retrieves a big.Int from the pool and sets it to the value of src
func (p *BigIntPool) GetCopy(src *big.Int) *big.Int {
	x := p.Get()
	x.Set(src)
	return x
}

// Global pool instance for convenience
var GlobalBigIntPool = NewBigIntPool()

// ByteSlicePool provides a pool of reusable byte slices to reduce allocations
type ByteSlicePool struct {
	pool sync.Pool
	size int
}

// NewByteSlicePool creates a new ByteSlicePool with fixed slice size
func NewByteSlicePool(size int) *ByteSlicePool {
	return &ByteSlicePool{
		size: size,
		pool: sync.Pool{
			New: func() interface{} {
				return make([]byte, size)
			},
		},
	}
}

// Get retrieves a byte slice from the pool
func (p *ByteSlicePool) Get() []byte {
	return p.pool.Get().([]byte)
}

// Put returns a byte slice to the pool after clearing it
func (p *ByteSlicePool) Put(b []byte) {
	if b != nil && len(b) == p.size {
		// Clear the slice
		for i := range b {
			b[i] = 0
		}
		p.pool.Put(b)
	}
}

// Global 32-byte slice pool for hash operations
var Global32BytePool = NewByteSlicePool(32)

// StringSlicePool provides a pool of reusable string slices
type StringSlicePool struct {
	pool sync.Pool
	size int
}

// NewStringSlicePool creates a new StringSlicePool with initial capacity
func NewStringSlicePool(size int) *StringSlicePool {
	return &StringSlicePool{
		size: size,
		pool: sync.Pool{
			New: func() interface{} {
				return make([]string, 0, size)
			},
		},
	}
}

// Get retrieves a string slice from the pool
func (p *StringSlicePool) Get() []string {
	return p.pool.Get().([]string)[:0] // Reset length but keep capacity
}

// Put returns a string slice to the pool
func (p *StringSlicePool) Put(s []string) {
	if s != nil && cap(s) >= p.size {
		p.pool.Put(s)
	}
}

// Global string slice pool for siblings arrays
var GlobalStringSlicePool = NewStringSlicePool(256) // Max tree depth

// InterfaceSlicePool provides a pool of reusable interface{} slices for database entries
type InterfaceSlicePool struct {
	pool sync.Pool
	size int
}

// NewInterfaceSlicePool creates a new InterfaceSlicePool with initial capacity
func NewInterfaceSlicePool(size int) *InterfaceSlicePool {
	return &InterfaceSlicePool{
		size: size,
		pool: sync.Pool{
			New: func() interface{} {
				return make([]interface{}, 0, size)
			},
		},
	}
}

// Get retrieves an interface{} slice from the pool
func (p *InterfaceSlicePool) Get() []interface{} {
	return p.pool.Get().([]interface{})[:0] // Reset length but keep capacity
}

// Put returns an interface{} slice to the pool
func (p *InterfaceSlicePool) Put(s []interface{}) {
	if s != nil && cap(s) >= p.size {
		// Clear references to prevent memory leaks
		for i := range s[:cap(s)] {
			s[i] = nil
		}
		p.pool.Put(s)
	}
}

// Global interface slice pool for database entries
var GlobalInterfaceSlicePool = NewInterfaceSlicePool(3) // Max 3 elements per database entry
