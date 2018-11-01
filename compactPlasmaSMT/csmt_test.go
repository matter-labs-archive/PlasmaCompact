package compactplasmasmt

import (
	"bytes"
	"log"
	"testing"
)

func TestInsertion(t *testing.T) {
	cache := make(CacheBranch)
	csmt := new(CSMT)
	csmt.cache = &cache
	csmt.Height = 4
	csmtLevel := new(CSMTLevel)
	csmtLevel.cache = &cache
	csmtLevel.MaxLevel = 4
	csmt.Root = csmtLevel
	toInsert := make(InsertionIndexes, 1)
	toInsert[0].Index = 0
	toInsert[0].Value = []byte{0x01}
	path := csmt.ApplyInserts(toInsert)
	// log.Println(path)
	valueHash := LeafHash([]byte{0x01})
	log.Println(valueHash)
	for i := 0; i < 4; i++ {
		valueHash = NodeHash(valueHash, nil)
		// log.Println(valueHash)
	}
	if bytes.Compare(valueHash, path[0].Value) != 0 {
		t.Fail()
	}
}

func TestInsertionAndDeletion(t *testing.T) {
	cache := make(CacheBranch)
	csmt := new(CSMT)
	csmt.cache = &cache
	csmt.Height = 4
	csmtLevel := new(CSMTLevel)
	csmtLevel.cache = &cache
	csmtLevel.MaxLevel = 4
	csmt.Root = csmtLevel
	toInsert := make(InsertionIndexes, 1)
	toInsert[0].Index = 0
	toInsert[0].Value = []byte{0x01}
	_ = csmt.ApplyInserts(toInsert)
	toDelete := make(DeletionIndexes, 1)
	toDelete[0] = 0
	path := csmt.ApplyDeletes(toDelete)
	for i := 0; i < len(path); i++ {
		if path[i].Value != nil {
			t.Fail()
		}
	}
}

func TestMultipleInsertion(t *testing.T) {
	cache := make(CacheBranch)
	csmt := new(CSMT)
	csmt.cache = &cache
	csmt.Height = 4
	csmtLevel := new(CSMTLevel)
	csmtLevel.cache = &cache
	csmtLevel.MaxLevel = 4
	csmt.Root = csmtLevel
	toInsert := make(InsertionIndexes, 2)
	toInsert[0].Index = 0
	toInsert[0].Value = []byte{0x01}
	toInsert[1].Index = 15
	toInsert[1].Value = []byte{0x02}
	path := csmt.ApplyInserts(toInsert)
	if len(path) != 9 {
		t.Fail()
	}
}

func TestPathVeficiationSmallHeight(t *testing.T) {
	cache := make(CacheBranch)
	csmt := new(CSMT)
	csmt.cache = &cache
	csmt.Height = 2
	csmtLevel := new(CSMTLevel)
	csmtLevel.cache = &cache
	csmtLevel.MaxLevel = 2
	csmt.Root = csmtLevel
	toInsert := make(InsertionIndexes, 1)
	toInsert[0].Index = 0
	toInsert[0].Value = []byte{0x01}
	path := csmt.ApplyInserts(toInsert)
	err := path.VefiryPath(2, 0, []byte{0x01}, path[0].Value)
	if err != nil {
		t.Fail()
	}
}

func TestPathVeficiation(t *testing.T) {
	cache := make(CacheBranch)
	csmt := new(CSMT)
	csmt.cache = &cache
	csmt.Height = 4
	csmtLevel := new(CSMTLevel)
	csmtLevel.cache = &cache
	csmtLevel.MaxLevel = 4
	csmt.Root = csmtLevel
	toInsert := make(InsertionIndexes, 1)
	toInsert[0].Index = 0
	toInsert[0].Value = []byte{0x01}
	path := csmt.ApplyInserts(toInsert)
	err := path.VefiryPath(4, 0, []byte{0x01}, path[0].Value)
	if err != nil {
		t.Fail()
	}
}

func TestPathVeficiationForMultiinsertSmallDepth(t *testing.T) {
	cache := make(CacheBranch)
	csmt := new(CSMT)
	csmt.cache = &cache
	csmt.Height = 2
	csmtLevel := new(CSMTLevel)
	csmtLevel.cache = &cache
	csmtLevel.MaxLevel = 2
	csmt.Root = csmtLevel
	toInsert := make(InsertionIndexes, 2)
	toInsert[0].Index = 0
	toInsert[0].Value = []byte{0x01}
	toInsert[1].Index = 3
	toInsert[1].Value = []byte{0x02}
	path := csmt.ApplyInserts(toInsert)
	filtered := path.FilterPath(2, 3)
	err := filtered.VefiryPath(2, 3, []byte{0x02}, path[0].Value)
	if err != nil {
		t.Fail()
	}
}

func TestPathVeficiationForMultiinsert(t *testing.T) {
	cache := make(CacheBranch)
	csmt := new(CSMT)
	csmt.cache = &cache
	csmt.Height = 4
	csmtLevel := new(CSMTLevel)
	csmtLevel.cache = &cache
	csmtLevel.MaxLevel = 4
	csmt.Root = csmtLevel
	toInsert := make(InsertionIndexes, 2)
	toInsert[0].Index = 0
	toInsert[0].Value = []byte{0x01}
	toInsert[1].Index = 15
	toInsert[1].Value = []byte{0x02}
	path := csmt.ApplyInserts(toInsert)
	filtered := path.FilterPath(4, 15)
	err := filtered.VefiryPath(4, 15, []byte{0x02}, path[0].Value)
	if err != nil {
		t.Fail()
	}
}

func TestInsertTwiceAndUpdate(t *testing.T) {
	cache := make(CacheBranch)
	csmt := new(CSMT)
	csmt.cache = &cache
	csmt.Height = 2
	csmtLevel := new(CSMTLevel)
	csmtLevel.cache = &cache
	csmtLevel.MaxLevel = 2
	csmt.Root = csmtLevel
	toInsert := make(InsertionIndexes, 1)
	toInsert[0].Index = 0
	toInsert[0].Value = []byte{0x01}
	toInsert2 := make(InsertionIndexes, 1)
	toInsert2[0].Index = 3
	toInsert2[0].Value = []byte{0x02}
	log.Println("Inserting the first node")
	path := csmt.ApplyInserts(toInsert)
	filtered := path.FilterPath(2, 0)
	log.Println("Inserting the second node")
	path2 := csmt.ApplyInserts(toInsert2)
	joined, err := filtered.UpdateProof(0, path2)
	if err != nil {
		t.Fatal("Failed to update a proof")
		return
	}
	if bytes.Compare(path2[0].Value, joined[0].Value) != 0 {
		t.Fatal("Joined root hash and original root hash did not match")
	}
	err = joined.VefiryPath(2, 0, []byte{0x01}, path2[0].Value)
	if err != nil {
		t.Fatal("Proof did not match")
	}
}

func TestInsertTwiceAndUpdateHigherDepth(t *testing.T) {
	cache := make(CacheBranch)
	csmt := new(CSMT)
	csmt.cache = &cache
	csmt.Height = 4
	csmtLevel := new(CSMTLevel)
	csmtLevel.cache = &cache
	csmtLevel.MaxLevel = 4
	csmt.Root = csmtLevel
	toInsert := make(InsertionIndexes, 1)
	toInsert[0].Index = 0
	toInsert[0].Value = []byte{0x01}
	toInsert2 := make(InsertionIndexes, 1)
	toInsert2[0].Index = 15
	toInsert2[0].Value = []byte{0x02}
	log.Println("Inserting the first node")
	path := csmt.ApplyInserts(toInsert)
	filtered := path.FilterPath(4, 0)
	log.Println("Inserting the second node")
	path2 := csmt.ApplyInserts(toInsert2)
	joined, err := filtered.UpdateProofImproved(0, path2)
	if err != nil {
		t.Fatal("Failed to update a proof")
		return
	}
	if bytes.Compare(path2[0].Value, joined[0].Value) != 0 {
		t.Fatal("Joined root hash and original root hash did not match")
	}
	err = joined.VefiryPath(4, 0, []byte{0x01}, path2[0].Value)
	if err != nil {
		t.Fatal("Proof did not match")
	}
}

func TestInsertTwiceAndUpdateDeepIntersectionLargeHeight(t *testing.T) {
	cache := make(CacheBranch)
	csmt := new(CSMT)
	csmt.cache = &cache
	csmt.Height = 4
	csmtLevel := new(CSMTLevel)
	csmtLevel.cache = &cache
	csmtLevel.MaxLevel = 4
	csmt.Root = csmtLevel
	toInsert := make(InsertionIndexes, 1)
	toInsert[0].Index = 0
	toInsert[0].Value = []byte{0x01}
	toInsert2 := make(InsertionIndexes, 1)
	toInsert2[0].Index = 2
	toInsert2[0].Value = []byte{0x02}
	log.Println("Inserting the first node")
	path := csmt.ApplyInserts(toInsert)
	cache.Print()
	filtered := path.FilterPath(4, 0)
	log.Println("Inserting the second node")
	path2 := csmt.ApplyInserts(toInsert2)
	cache.Print()
	joined, err := filtered.UpdateProof(0, path2)
	if err != nil {
		t.Fatal("Failed to update a proof")
		return
	}
	if bytes.Compare(path2[0].Value, joined[0].Value) != 0 {
		t.Fatal("Joined root hash and original root hash did not match")
	}
	err = joined.VefiryPath(4, 0, []byte{0x01}, path2[0].Value)
	if err != nil {
		t.Fatal("Proof did not match")
	}
}

func TestInsertTwiceAndUpdateDeepIntersection(t *testing.T) {
	cache := make(CacheBranch)
	csmt := new(CSMT)
	csmt.cache = &cache
	csmt.Height = 2
	csmtLevel := new(CSMTLevel)
	csmtLevel.cache = &cache
	csmtLevel.MaxLevel = 2
	csmt.Root = csmtLevel
	toInsert := make(InsertionIndexes, 1)
	toInsert[0].Index = 0
	toInsert[0].Value = []byte{0x01}
	toInsert2 := make(InsertionIndexes, 1)
	toInsert2[0].Index = 1
	toInsert2[0].Value = []byte{0x02}
	log.Println("Inserting the first node")
	path := csmt.ApplyInserts(toInsert)
	cache.Print()
	filtered := path.FilterPath(2, 0)
	log.Println("Inserting the second node")
	path2 := csmt.ApplyInserts(toInsert2)
	cache.Print()
	joined, err := filtered.UpdateProofImproved(0, path2)
	if err != nil {
		t.Fatal("Failed to update a proof")
		return
	}
	if bytes.Compare(path2[0].Value, joined[0].Value) != 0 {
		t.Fatal("Joined root hash and original root hash did not match")
	}
	err = joined.VefiryPath(2, 0, []byte{0x01}, path2[0].Value)
	if err != nil {
		t.Fatal("Proof did not match")
	}
}

func TestInsertTwiceWithMultipleInserts(t *testing.T) {
	cache := make(CacheBranch)
	csmt := new(CSMT)
	csmt.cache = &cache
	csmt.Height = 4
	csmtLevel := new(CSMTLevel)
	csmtLevel.cache = &cache
	csmtLevel.MaxLevel = 4
	csmt.Root = csmtLevel
	toInsert := make(InsertionIndexes, 1)
	toInsert[0].Index = 0
	toInsert[0].Value = []byte{0x01}
	toInsert2 := make(InsertionIndexes, 2)
	toInsert2[0].Index = 5
	toInsert2[0].Value = []byte{0x02}
	toInsert2[1].Index = 11
	toInsert2[1].Value = []byte{0x03}
	log.Println("Inserting the first node")
	path := csmt.ApplyInserts(toInsert)
	cache.Print()
	filtered := path.FilterPath(4, 0)
	log.Println("Inserting the second node")
	path2 := csmt.ApplyInserts(toInsert2)
	cache.Print()
	joined, err := filtered.UpdateProof(0, path2)
	if err != nil {
		t.Fatal("Failed to update a proof")
		return
	}
	if bytes.Compare(path2[0].Value, joined[0].Value) != 0 {
		t.Fatal("Joined root hash and original root hash did not match")
	}
	err = joined.VefiryPath(4, 0, []byte{0x01}, path2[0].Value)
	if err != nil {
		t.Fatal("Proof did not match")
	}
}

func TestInsertTwiceWithMultipleInsertsRewrite(t *testing.T) {
	cache := make(CacheBranch)
	csmt := new(CSMT)
	csmt.cache = &cache
	csmt.Height = 4
	csmtLevel := new(CSMTLevel)
	csmtLevel.cache = &cache
	csmtLevel.MaxLevel = 4
	csmt.Root = csmtLevel
	toInsert := make(InsertionIndexes, 1)
	toInsert[0].Index = 0
	toInsert[0].Value = []byte{0x01}
	toInsert2 := make(InsertionIndexes, 2)
	toInsert2[0].Index = 5
	toInsert2[0].Value = []byte{0x02}
	toInsert2[1].Index = 11
	toInsert2[1].Value = []byte{0x03}
	log.Println("Inserting the first node")
	path := csmt.ApplyInserts(toInsert)
	cache.Print()
	filtered := path.FilterPath(4, 0)
	log.Println("Inserting the second node")
	path2 := csmt.ApplyInserts(toInsert2)
	cache.Print()
	joined, err := filtered.UpdateProofImproved(0, path2)
	if err != nil {
		t.Fatal("Failed to update a proof")
		return
	}
	if bytes.Compare(path2[0].Value, joined[0].Value) != 0 {
		t.Fatal("Joined root hash and original root hash did not match")
	}
	err = joined.VefiryPath(4, 0, []byte{0x01}, path2[0].Value)
	if err != nil {
		t.Fatal("Proof did not match")
	}
}
