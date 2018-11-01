package compactplasmasmt

import (
	"encoding/binary"
	"log"
)

// CacheBranch caches every branch where both children have non-default values.
type CacheBranch map[string][]byte

// Exists checks if a value exists in the cache.
func (c CacheBranch) Exists(height uint8, nodeID uint64) bool {
	index := []byte{height}
	nodeIDbytes := make([]byte, 8)
	binary.BigEndian.PutUint64(nodeIDbytes, nodeID)
	index = append(index, nodeIDbytes...)
	_, exists := c[string(index)]
	return exists
}

// Get returns a value that exists from the cache.
func (c CacheBranch) Get(height uint8, nodeID uint64) []byte {
	index := []byte{height}
	nodeIDbytes := make([]byte, 8)
	binary.BigEndian.PutUint64(nodeIDbytes, nodeID)
	index = append(index, nodeIDbytes...)
	value := c[string(index)]
	return value
}

// HashCache hashes the provided values and maybe caches.
func (c CacheBranch) UpdateAndStore(height uint8, nodeID uint64, value []byte) []byte {
	newHash := value
	if newHash == nil {
		c.Delete(height, nodeID)
		return nil
	}
	index := []byte{height}
	nodeIDbytes := make([]byte, 8)
	binary.BigEndian.PutUint64(nodeIDbytes, nodeID)
	index = append(index, nodeIDbytes...)
	c[string(index)] = newHash
	return newHash
}

func (c CacheBranch) Delete(height uint8, nodeID uint64) bool {
	index := []byte{height}
	nodeIDbytes := make([]byte, 8)
	binary.BigEndian.PutUint64(nodeIDbytes, nodeID)
	index = append(index, nodeIDbytes...)
	_, exists := c[string(index)]
	delete(c, string(index))
	return exists
}

func (c CacheBranch) Insert(height uint8, nodeID uint64, value []byte) {
	index := []byte{height}
	nodeIDbytes := make([]byte, 8)
	binary.BigEndian.PutUint64(nodeIDbytes, nodeID)
	index = append(index, nodeIDbytes...)
	c[string(index)] = value
}

// Entries returns the number of entries in the cache.
func (c CacheBranch) Entries() int {
	return len(c)
}

func (c CacheBranch) Print() {
	for k, v := range c {
		log.Printf("For key %v value = %v", []byte(k), v)
	}
}
