/* slab_class_test.go - unit test for slab_class.go */
/*
modification history
--------------------
2014/12/3, by Sijie Yang, create
*/
/*
DESCRIPTION
*/
package slab_pool

import "testing"

func TestSlabClassGrowth(t *testing.T) {
    slabClass := NewSlabClass(4096, 2048, 201412)
    if len(slabClass.slabs) != 0 {
        t.Errorf("count for slabs in slabClass should be 0")
    }

    // allocate one slab
    chunk, _ := slabClass.chunkAlloc()
    if len(chunk) != 2048 {
        t.Errorf("chunk size should be 2048")
    }
    if len(slabClass.slabs) != 1 {
        t.Errorf("count for slabs in slabClass should be 1")
    }

    // allocate another slab
    chunk, _ = slabClass.chunkAlloc()
    chunk, _ = slabClass.chunkAlloc()
    if len(slabClass.slabs) != 2 {
        t.Errorf("count for slabs in slabClass should be 2")
    }
}

func TestSlabTransfer(t *testing.T) {
    // prepare slab class
    slabPool, _ := CreateSlabPool(4096, 1024, 2048, 2)
    chunk1, _ := slabPool.Get(2048)
    slab1, chunkIndex1, _ := slabPool.locate(chunk1)
    slabClass := slab1.slabClass

    // slabLists checker
    check := func(isEmptyFree, isEmpytUse, isEmptyFull bool) {
        if isEmptyFree != slabClass.listEmpty(SLAB_FREE) {
            t.Errorf("slabFree should free: %t, got %t", isEmptyFree, !isEmptyFree)
        }
        if isEmpytUse != slabClass.listEmpty(SLAB_USE) {
            t.Errorf("slabFree should free: %t, got %t", isEmpytUse, !isEmpytUse)
        }
        if isEmptyFull != slabClass.listEmpty(SLAB_FULL) {
            t.Errorf("slabFree should free: %t, got %t", isEmptyFull, !isEmptyFull)
        }
    }

    // one slab with one free chunk
    if len(slabClass.slabs) != 1 {
        t.Errorf("count for slabs in slabClass should be 1")
    }
    check(true, false, true)

    // one slab with no free chunk
    chunk2, _ := slabClass.chunkAlloc()
    slab2, chunkIndex2, _ := slabPool.locate(chunk2)
    if slab1 != slab2 {
        t.Errorf("not alloc chunks from SLAB_USE")
    }
    check(true, true, false)

    // one slab from which no chunk is allocted
    slabClass.chunkDecRef(slab1, chunkIndex1)
    slabClass.chunkDecRef(slab2, chunkIndex2)
    check(false, true, true)

    // allocate from slabFree
    slabClass.chunkAlloc()
    check(true, false, true)
}

func TestSpeicalSlab(t *testing.T) {
    // prepare slab class
    slabPool, _ := CreateSlabPool(4096, 4096, 4096, 2)
    chunk, _ := slabPool.Get(4096)
    slab, chunkIndex, _ := slabPool.locate(chunk)
    slabClass := slab.slabClass

    // check chunk
    if chunk == nil || len(chunk) != slabClass.slabSize {
        t.Errorf("shuold return valid chunk with same size as slab")
    }

    slabClass.chunkDecRef(slab, chunkIndex)

    // check chunk
    chunk, _ = slabClass.chunkAlloc()
    if chunk == nil || len(chunk) != slabClass.slabSize {
        t.Errorf("shuold return valid chunk with same size as slab")
    }
}

func TestSlabListOperation(t *testing.T) {
    // prepare slab class
    slabPool, _ := CreateSlabPool(4096, 4096, 4096, 2)
    chunk, _ := slabPool.Get(4096)
    slab, _, _ := slabPool.locate(chunk)
    slabClass := slab.slabClass

    // three slab in slabFull list
    chunk1, _ := slabPool.Get(4096)
    chunk2, _ := slabPool.Get(4096)

    slab1, chunkIndex1, _ := slabPool.locate(chunk1)
    slab2, chunkIndex2, _ := slabPool.locate(chunk2)

    // remove the middle node from slabFull list
    slabClass.chunkDecRef(slab1, chunkIndex1)
    if slabClass.slabLists[SLAB_FULL] < 0 {
        t.Errorf("size of SlabFull should not be 1")
    }

    // remove the first node from slabFull list
    slabClass.chunkDecRef(slab2, chunkIndex2)
    if slabClass.slabLists[SLAB_FULL] < 0 {
        t.Errorf("size of SlabFull should not be 1")
    }
}

