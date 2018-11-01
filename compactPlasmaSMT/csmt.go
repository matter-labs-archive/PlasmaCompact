package compactplasmasmt

import (
	"bytes"
	"crypto/sha512"
	"errors"
	"log"
	"sort"
)

const (
	treeHeight            = 24 + 20 + 4 // maximum 2^24 blocks, 2^20 transactions per block, 2^4 inputs/outputs per TX
	operationDelete       = uint8(1)
	operationAdd          = uint8(2)
	blockPrefixBits       = 24
	transactionPrefixBits = 20
	outputPrefixBits      = 4
)

func NodeHash(left, right []byte) []byte {
	if left == nil && right == nil {
		return nil
	} else if right == nil {
		hasher := sha512.New512_256()
		hasher.Write(left)
		return hasher.Sum(nil)
	} else if left == nil {
		hasher := sha512.New512_256()
		hasher.Write(right)
		return hasher.Sum(nil)
	}
	hasher := sha512.New512_256()
	hasher.Write(left)
	hasher.Write(right)
	return hasher.Sum(nil)
}

func LeafHash(leaf []byte) []byte {
	hasher := sha512.New512_256()
	hasher.Write(leaf)
	return hasher.Sum(nil)
}

type InsertedIndex struct {
	Index uint64
	Value []byte
}

type InsertionIndexes []InsertedIndex

// sort.Interface method for sorting
func (d InsertionIndexes) Len() int           { return len(d) }
func (d InsertionIndexes) Swap(i, j int)      { d[i], d[j] = d[j], d[i] }
func (d InsertionIndexes) Less(i, j int) bool { return d[i].Index < d[j].Index }
func (d InsertionIndexes) Split(bit uint8) (l, r InsertionIndexes) {
	// // can replace it with binary search
	// log.Printf("Split bit is %v", bit)
	// log.Println("Indexes to split:")
	// for j := 0; j < len(d); j++ {
	// 	log.Printf("%b", d[j].Index)
	// }
	mask := (uint64(1) << bit) - 1
	// log.Printf("Mask = %b", mask)
	splitValue := uint64(1) << (bit - 1)
	// log.Printf("Split value = %b", splitValue)
	// log.Printf("Masked values")
	// for j := 0; j < len(d); j++ {
	// log.Printf("%b", d[j].Index&mask)
	// }
	i := sort.Search(len(d), func(k int) bool {
		maskedValue := (d[k].Index & mask)
		return maskedValue >= splitValue
	})
	return d[:i], d[i:]
}

type DeletionIndexes []uint64

// sort.Interface method for sorting
func (d DeletionIndexes) Len() int           { return len(d) }
func (d DeletionIndexes) Swap(i, j int)      { d[i], d[j] = d[j], d[i] }
func (d DeletionIndexes) Less(i, j int) bool { return d[i] < d[j] }
func (d DeletionIndexes) Split(bit uint8) (l, r DeletionIndexes) {
	// can replace it with binary search
	// log.Println("Indexes to split:")
	// for j := 0; j < len(d); j++ {
	// 	log.Printf("%b", d[j])
	// }
	mask := (uint64(1) << bit) - 1
	// log.Printf("Mask = %b", mask)
	splitValue := uint64(1) << (bit - 1)
	// log.Printf("Split bit = %b", splitValue)
	i := sort.Search(d.Len(), func(i int) bool {
		return (d[i] & mask) >= splitValue
	})
	return d[:i], d[i:]
}

type AuditNode struct {
	Level        uint8
	Index        uint64
	Value        []byte
	LeftSibling  []byte
	RightSibling []byte
}

type AuditNodes []AuditNode

func (d AuditNodes) Len() int           { return len(d) }
func (d AuditNodes) Swap(i, j int)      { d[i], d[j] = d[j], d[i] }
func (d AuditNodes) Less(i, j int) bool { return d[i].Index < d[j].Index }

// SMT is a sparse Merkle tree.
type CSMT struct {
	cache  *CacheBranch // Cache interface could be implemented by different caching strategies
	Height uint8        // key of left-most leaf of a subtree, fixed in size.
	Root   *CSMTLevel
}

type CSMTLevel struct {
	cache    *CacheBranch // Cache interface could be implemented by different caching strategies
	MaxLevel uint8        // level in the global tree, bottom level == 0
}

// Indexes to delete must be always sorted and only point to the bottom of the tree.
func (s *CSMT) ApplyDeletes(d DeletionIndexes) AuditNodes {
	return s.Root.ApplyDeletes(d, s.Height)
}

func (s *CSMTLevel) ApplyDeletes(d DeletionIndexes, splitLevel uint8) AuditNodes {
	if len(d) == 0 {
		return nil
	}
	lastLevel := splitLevel == 0
	if lastLevel {
		if len(d) != 1 {
			log.Fatalln("Trying to delete not a single indexes at the bottom level")
		}
		_ = s.cache.Delete(splitLevel, d[0])
		node := make(AuditNodes, 1)
		node[0].Level = 0
		node[0].Index = d[0]
		node[0].Value = nil
		return node
	}
	l, r := d.Split(splitLevel)
	left := s.ApplyDeletes(l, splitLevel-1)
	right := s.ApplyDeletes(r, splitLevel-1)

	lenLeft := len(left)
	lenRight := len(right)

	var leftRoot AuditNode
	if lenLeft != 0 {
		leftRoot = left[0]
	}

	var rightRoot AuditNode
	if lenRight != 0 {
		rightRoot = right[0]
	}
	thisHash := NodeHash(leftRoot.Value, rightRoot.Value)
	var thisID uint64
	if lenLeft != 0 {
		thisID = leftRoot.Index >> 1
	} else if lenRight != 0 {
		thisID = rightRoot.Index >> 1
	} else {
		log.Panicln("Can not get audit node index")
	}
	s.cache.Delete(splitLevel, thisID)

	allNodes := make(AuditNodes, lenLeft+lenRight+1)
	allNodes[0] = AuditNode{splitLevel, thisID, thisHash, leftRoot.Value, rightRoot.Value}
	for i := 0; i < lenLeft; i++ {
		allNodes[1+i] = left[i]
	}
	for i := 0; i < lenRight; i++ {
		allNodes[1+lenLeft+i] = right[i]
	}
	return allNodes
}

func (s *CSMT) ApplyInserts(d InsertionIndexes) AuditNodes {
	return s.Root.ApplyInserts(d, s.Height)
}

func (s *CSMTLevel) ApplyInserts(d InsertionIndexes, splitLevel uint8) AuditNodes {
	if len(d) == 0 {
		// node := make(AuditNodes, 1)
		// node[0].Level = splitLevel
		// node[0].Index = 0
		// node[0].Value = nil
		// return node
		return nil
	}
	lastLevel := splitLevel == 0
	if lastLevel {
		if len(d) != 1 {
			log.Fatalln("Trying to insert not a single indexes at the bottom level")
		}
		newHash := LeafHash(d[0].Value)
		s.cache.Insert(splitLevel, d[0].Index, newHash)
		node := make(AuditNodes, 1)
		node[0].Level = 0
		node[0].Index = d[0].Index
		node[0].Value = newHash
		return node
	}
	l, r := d.Split(splitLevel)
	left := s.ApplyInserts(l, splitLevel-1)
	right := s.ApplyInserts(r, splitLevel-1)

	lenLeft := len(left)
	lenRight := len(right)

	if lenLeft == 0 && lenRight == 0 {
		return nil
	}

	var leftRoot AuditNode
	var rightRoot AuditNode
	if lenLeft != 0 && lenRight == 0 {
		leftRoot = left[0]
		cacheRecord := s.cache.Get(leftRoot.Level, leftRoot.Index+1)
		if cacheRecord != nil {
			n := AuditNode{leftRoot.Level, leftRoot.Index + 1, cacheRecord, nil, nil}
			rightRoot = n
		}
	} else if lenLeft == 0 && lenRight != 0 {
		rightRoot = right[0]
		cacheRecord := s.cache.Get(rightRoot.Level, rightRoot.Index-1)
		if cacheRecord != nil {
			n := AuditNode{rightRoot.Level, rightRoot.Index - 1, cacheRecord, nil, nil}
			leftRoot = n
		}
	} else {
		leftRoot = left[0]
		rightRoot = right[0]
	}

	thisHash := NodeHash(leftRoot.Value, rightRoot.Value)
	var thisID uint64
	if lenLeft != 0 {
		thisID = leftRoot.Index >> 1
	} else if lenRight != 0 {
		thisID = rightRoot.Index >> 1
	} else {
		log.Panicln("Can not get audit node index")
	}
	s.cache.UpdateAndStore(splitLevel, thisID, thisHash)

	allNodes := make(AuditNodes, lenLeft+lenRight+1)
	allNodes[0] = AuditNode{splitLevel, thisID, thisHash, leftRoot.Value, rightRoot.Value}
	for i := 0; i < lenLeft; i++ {
		allNodes[1+i] = left[i]
	}
	for i := 0; i < lenRight; i++ {
		allNodes[1+lenLeft+i] = right[i]
	}
	return allNodes
}

func (p AuditNodes) VefiryPath(height uint8, index uint64, value, root []byte) error {
	if len(p) == 0 {
		return errors.New("Path can not be zero length")
	}
	if uint8(len(p)-1) != height {
		return errors.New("Path length is invalid")
	}
	if bytes.Compare(p[0].Value, root) != 0 {
		return errors.New("Low level hash does not match")
	}
	if p[len(p)-1].Index != index {
		return errors.New("Most likely checking for invalid index")
	}
	hash := LeafHash(value)
	if bytes.Compare(p[len(p)-1].Value, hash) != 0 {
		return errors.New("Low level hash does not match")
	}
	idx := index
	for i := len(p) - 2; i >= 0; i-- {
		thisP := p[i]
		if idx&1 == 0 {
			proof := thisP.RightSibling
			hash = NodeHash(hash, proof)
		} else {
			proof := thisP.LeftSibling
			hash = NodeHash(proof, hash)
		}
		idx = idx / 2
	}
	if bytes.Compare(root, hash) != 0 {
		return errors.New("Audit path failed")
	}
	return nil
}

func (p AuditNodes) FilterPath(height uint8, index uint64) AuditNodes {
	filtered := make(AuditNodes, height+1)
	maxHops := len(p)
	j := uint64(height)
	idx := index
	expectedLevel := uint8(0)
	for i := maxHops - 1; i >= 0; i-- {
		if p[i].Index == idx {
			thisP := p[i]
			if expectedLevel != thisP.Level {
				continue
				// log.Fatalln("Filtering excountered unexpected level")
			}
			expectedLevel++
			idx = idx / 2
			filtered[j] = AuditNode{thisP.Level, thisP.Index, thisP.Value, thisP.LeftSibling, thisP.RightSibling}
			if j == 0 {
				break
			}
			j--
		}
	}
	return filtered
}

func (p AuditNodes) UpdateProof(index uint64, extraData AuditNodes) (AuditNodes, error) {
	joined := make(AuditNodes, len(p))
	maxHops := len(extraData)

	expectedLevel := uint8(0)

	// first - just copy the proof
	for i := 0; i < len(p); i++ {
		joined[i] = p[i]
	}
	joiningIndex := int(p[0].Level)
	joiningIndex--

	firstIntersectionFound := false
	nodeIntersectionFound := false
	intersectionIndex := joiningIndex
	for j := joiningIndex; j >= 0; j-- {
		currentNode := joined[j]
		siblingLeft := currentNode.Index * 2
		siblingRight := siblingLeft + 1
		expectedLevel = currentNode.Level - 1
		expectingLeft := currentNode.RightSibling != nil
		nodeIntersectionFound = false
		log.Printf("Updating node number %v at level %v", currentNode.Index, currentNode.Level)
		if expectingLeft {
			log.Printf("Looking for siblings with number %v at level %v", siblingLeft, expectedLevel)
		} else {
			log.Printf("Looking for siblings with number %v at level %v", siblingRight, expectedLevel)
		}
		for i := maxHops - 1; i >= 0; i-- {
			extraNode := extraData[i]
			if expectedLevel == extraNode.Level && extraNode.Index == siblingLeft && expectingLeft {
				if expectedLevel == 0 {
					if siblingLeft == index {
						return nil, errors.New("Self-update is forbidden, use filter instead")
					}
				}
				log.Printf("Found siblings with number %v at level %v", siblingLeft, expectedLevel)
				newValue := NodeHash(extraNode.Value, currentNode.RightSibling)
				joined[j] = AuditNode{currentNode.Level, currentNode.Index, newValue, extraNode.Value, currentNode.RightSibling}
				maxHops = i + 1
				firstIntersectionFound = true
				nodeIntersectionFound = true
				intersectionIndex = j
				break
			} else if expectedLevel == extraNode.Level && extraNode.Index == siblingRight && !expectingLeft {
				if expectedLevel == 0 {
					if siblingRight == index {
						return nil, errors.New("Self-update is forbidden, use filter instead")
					}
				}
				log.Printf("Found siblings with number %v at level %v", siblingRight, expectedLevel)
				newValue := NodeHash(currentNode.LeftSibling, extraNode.Value)
				joined[j] = AuditNode{currentNode.Level, currentNode.Index, newValue, currentNode.LeftSibling, extraNode.Value}
				maxHops = i + 1
				firstIntersectionFound = true
				nodeIntersectionFound = true
				intersectionIndex = j
				break
			}
		}
		if !nodeIntersectionFound {
			log.Println("Did not found a sibling in proof update data")
			if !firstIntersectionFound {
				maxHops = len(extraData)
			} else {
				if j != 0 {
					log.Println("There is no proof, and this level node was already updated, so update has to be propageted to higher level")
					// we are not on the last level yet, so can update and continue
					nextLevelNode := joined[j-1]
					replaceLeft := currentNode.Index&1 == 0
					if replaceLeft {
						newValue := NodeHash(currentNode.Value, nextLevelNode.RightSibling)
						joined[j-1] = AuditNode{nextLevelNode.Level, nextLevelNode.Index, newValue, currentNode.Value, nextLevelNode.RightSibling}
					} else {
						newValue := NodeHash(nextLevelNode.LeftSibling, currentNode.Value)
						joined[j-1] = AuditNode{nextLevelNode.Level, nextLevelNode.Index, newValue, nextLevelNode.LeftSibling, currentNode.Value}
					}
				}
			}
		} else {
			intersectionNode := joined[intersectionIndex]
			levelOfIntersection := intersectionNode.Level
			log.Printf("Intersection was found at node number %v at level %v", intersectionNode.Index, levelOfIntersection)
			if int(levelOfIntersection) != len(joined) && j != 0 {
				log.Println("Update to be propagated to higher level after intersection")
				// we are not on the last level yet, so can update and continue
				nextLevelNode := joined[j-1]
				replaceLeft := intersectionNode.Index&1 == 0
				if replaceLeft {
					log.Printf("Replacing left sibling of node number %v at level %v", nextLevelNode.Index, nextLevelNode.Level)
					newValue := NodeHash(intersectionNode.Value, nextLevelNode.RightSibling)
					joined[j-1] = AuditNode{nextLevelNode.Level, nextLevelNode.Index, newValue, intersectionNode.Value, nextLevelNode.RightSibling}
				} else {
					log.Printf("Replacing right sibling of node number %v at level %v", nextLevelNode.Index, nextLevelNode.Level)
					newValue := NodeHash(nextLevelNode.LeftSibling, intersectionNode.Value)
					joined[j-1] = AuditNode{nextLevelNode.Level, nextLevelNode.Index, newValue, nextLevelNode.LeftSibling, intersectionNode.Value}
				}
			}
		}
	}
	if !firstIntersectionFound {
		log.Println("No intersection until root at least")
		thisRoot := p[0]
		otherRoot := extraData[0]
		if thisRoot.LeftSibling != nil && otherRoot.RightSibling != nil {
			newValue := NodeHash(thisRoot.LeftSibling, otherRoot.RightSibling)
			joined[0] = AuditNode{0, 0, newValue, thisRoot.LeftSibling, otherRoot.RightSibling}
		} else if thisRoot.RightSibling != nil && otherRoot.LeftSibling != nil {
			newValue := NodeHash(otherRoot.LeftSibling, thisRoot.RightSibling)
			joined[0] = AuditNode{0, 0, newValue, otherRoot.LeftSibling, thisRoot.RightSibling}
		} else {
			return nil, errors.New("Unexpected intersection at root")
		}
		for i := 1; i < len(p); i++ {
			joined[i] = p[i]
		}
	}
	return joined, nil
}

func (p AuditNodes) UpdateProofImproved(index uint64, extraData AuditNodes) (AuditNodes, error) {
	joined := make(AuditNodes, len(p))
	maxHops := len(extraData)
	// first - just copy old proof
	for i := 0; i < len(p); i++ {
		joined[i] = p[i]
	}
	joiningIndex := int(p[0].Level) - 1
	nodeIntersectionFound := false
	for j := joiningIndex; j >= 0; j-- {
		currentNode := joined[j]
		previousNode := joined[j+1]
		currentNodePredecessorIsLeft := previousNode.Index&1 == 0
		if !nodeIntersectionFound {
			for i := maxHops - 1; i >= 0; i-- {
				extraNode := extraData[i]
				if extraNode.Level == currentNode.Level && extraNode.Index == currentNode.Index {
					log.Println("Proof should be updated")
					if currentNodePredecessorIsLeft {
						if bytes.Compare(extraNode.LeftSibling, currentNode.LeftSibling) != 0 {
							return nil, errors.New("subbranches has diverged")
						}
					} else {
						if bytes.Compare(extraNode.RightSibling, currentNode.RightSibling) != 0 {
							return nil, errors.New("subbranches has diverged")
						}
					}
					joined[j] = extraNode
					// if currentNodePredecessorIsLeft {
					// 	joined[j].RightSibling = nil
					// } else {
					// 	joined[j].LeftSibling = nil
					// }
					log.Println("Interseciton found")
					nodeIntersectionFound = true
					maxHops = i
					break
				}
			}
			if !nodeIntersectionFound {
				maxHops = len(extraData)
			}
		} else {
			for i := maxHops - 1; i >= 0; i-- {
				extraNode := extraData[i]
				if extraNode.Level == currentNode.Level && extraNode.Index == currentNode.Index {
					if currentNodePredecessorIsLeft {
						if bytes.Compare(extraNode.LeftSibling, previousNode.Value) != 0 {
							return nil, errors.New("subbranches has diverged")
						}
					} else {
						if bytes.Compare(extraNode.RightSibling, previousNode.Value) != 0 {
							return nil, errors.New("subbranches has diverged")
						}
					}
					joined[j] = extraNode
					// if currentNodePredecessorIsLeft {
					// 	joined[j].RightSibling = nil
					// } else {
					// 	joined[j].LeftSibling = nil
					// }
					maxHops = i
					break
				}
			}
		}
	}
	return joined, nil
}
