package compactplasmasmt

import (
	"bytes"
	"log"
	"math/rand"
	"sort"
	"testing"
	"time"
)

const (
	totalPlasmaHeight = 48
	blockNumberLength = 24
)

// const (
// 	totalPlasmaHeight = 8
// 	blockNumberLength = 4
// )

// creepy random
func random(min, max uint64) uint64 {
	return (rand.Uint64() % (max - min)) + min
}

const numToInsert = 100000

func TestKindOfPlasma(t *testing.T) {
	rand.Seed(time.Now().Unix())
	cache := make(CacheBranch)
	csmt := new(CSMT)
	csmt.cache = &cache
	csmt.Height = totalPlasmaHeight
	csmtLevel := new(CSMTLevel)
	csmtLevel.cache = &cache
	csmtLevel.MaxLevel = totalPlasmaHeight
	csmt.Root = csmtLevel
	blockNumber := uint64(1) << blockNumberLength
	maxOutputNumber := uint64(1) << (totalPlasmaHeight - blockNumberLength)
	log.Println("Producing block 1")
	toInsert := make(InsertionIndexes, 0)
	allIndexes := make(map[uint64]bool)
	for i := 0; i < numToInsert; i++ {
		randomBytes := make([]byte, 32+64) // amount + pub key
		rand.Read(randomBytes)
		randomOutputNumberInBlock := random(0, maxOutputNumber)
		indexNumber := randomOutputNumberInBlock + blockNumber
		_, exists := allIndexes[indexNumber]
		if exists == true {
			continue
		}
		// log.Printf("Inserting at index %v", randomOutputNumberInBlock+blockNumber)
		newValue := new(InsertedIndex)
		newValue.Index = indexNumber
		newValue.Value = randomBytes
		toInsert = append(toInsert, *newValue)
		allIndexes[indexNumber] = true
	}
	// SORT! insersion indexes
	sort.Sort(&toInsert)
	now := time.Now()
	path := csmt.ApplyInserts(toInsert)
	elapsed := time.Since(now)
	log.Printf("Insertion of %v entities for tree of height %v has taken %v ms", len(toInsert), totalPlasmaHeight, float64(elapsed.Nanoseconds())/float64(1000000.0))
	log.Printf("Modeled TPS = %v", float64(len(toInsert))/float64(elapsed.Nanoseconds())*float64(1000000000))
	randInt := rand.Intn(len(toInsert))
	ourOutput := toInsert[randInt]
	log.Printf("Our output is number %v", ourOutput.Index)
	filtered := path.FilterPath(uint8(totalPlasmaHeight), ourOutput.Index)
	log.Printf("Total auxilary block information for proof updates will be roughtly %v kB", len(path)*33/1024)

	log.Println("Producing block 2")
	blockNumber = uint64(2) << blockNumberLength
	toInsert = make(InsertionIndexes, 0)
	for i := 0; i < numToInsert; i++ {
		randomBytes := make([]byte, 32+64) // amount + pub key
		rand.Read(randomBytes)
		randomOutputNumberInBlock := random(0, maxOutputNumber)
		indexNumber := randomOutputNumberInBlock + blockNumber
		_, exists := allIndexes[indexNumber]
		if exists == true {
			continue
		}
		// log.Printf("Inserting at index %v", randomOutputNumberInBlock+blockNumber)
		newValue := new(InsertedIndex)
		newValue.Index = indexNumber
		newValue.Value = randomBytes
		toInsert = append(toInsert, *newValue)
		allIndexes[indexNumber] = true
	}
	sort.Sort(&toInsert)
	now = time.Now()
	path2 := csmt.ApplyInserts(toInsert)
	elapsed = time.Since(now)
	log.Printf("Insertion of %v entities for tree of height %v has taken %v ms", len(toInsert), totalPlasmaHeight, float64(elapsed.Nanoseconds())/float64(1000000.0))
	log.Printf("Modeled TPS = %v", float64(len(toInsert))/float64(elapsed.Nanoseconds())*float64(1000000000))

	log.Printf("Total auxilary block information for proof updates will be roughtly %v kB", len(path2)*33/1024)

	joined, err := filtered.UpdateProofImproved(ourOutput.Index, path2)
	if err != nil {
		t.Fatal("Failed to update a proof")
		return
	}
	if bytes.Compare(path2[0].Value, joined[0].Value) != 0 {
		t.Fatal("Joined root hash and original root hash did not match")
	}
	newRoot := path2[0].Value
	err = joined.VefiryPath(totalPlasmaHeight, ourOutput.Index, ourOutput.Value, newRoot)
	if err != nil {
		t.Fatal("Proof did not match")
	}
}
