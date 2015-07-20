/* slab_pool_test.go - unit test for slab_pool.go */
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

func TestCreateSlabPool(t *testing.T) {
    var err error
    _, err = CreateSlabPool(4096, 0, 32, 1.2)
    if err == nil {
        t.Errorf("expected error return due to wrong chunkSizeMin")
    }
    _, err = CreateSlabPool(4096, 32, 16, 1.2)
    if err == nil {
        t.Errorf("expected error return due to wrong chunkSizeMin/Max")
    }
    _, err = CreateSlabPool(24, 16, 32, 1.2)
    if err == nil {
        t.Errorf("expected error return due to wrong slabSize")
    }
    _, err = CreateSlabPool(32, 8, 16, 0.8)
    if err == nil {
        t.Errorf("expected error return due to wrong fractor")
    }
    _, err = CreateSlabPool(32, 8, 16, 1.1)
    if err == nil {
        t.Errorf("expected error return due to wrong fractor")
    }
    _, err = CreateSlabPool(4096, 32, 64, 2)
    if err != nil {
        t.Errorf("expected slabpool return")
    }
}

func TestSlabClassFor(t *testing.T) {
    slabPool, _ := CreateSlabPool(4096, 64, 1024, 2)

    // test func: compare chunkSize of matched slabClass with expected chunkSize
    test := func (bufSize int, chunkSize int) {
        slabClass := slabPool.slabClassFor(bufSize)
        if slabClass.chunkSize != chunkSize {
            t.Errorf("expected slabClass with chunkSize:%s, got:%s", 
                     chunkSize, slabClass.chunkSize)
        }
    }

    // test cases
    test(64, 64)
    test(65, 128)
    test(127, 128)
    test(128, 128)
    test(256, 256)
    test(257, 512)
    test(512, 512)
    test(1023, 1024)
    test(1024, 1024)
}

func TestGet(t *testing.T) {
    slabPool, _ := CreateSlabPool(4096, 64, 1024, 2)

    testErr := func(size int) {
        _, err := slabPool.Get(size)
        if err == nil {
            t.Error("expected error due to wrong params")
        }
    }

    testOK := func(size int) {
        chunk, _ := slabPool.Get(size)
        if chunk == nil || len(chunk) != size {
            t.Error("should return valid chunk")
        }
        slab, chunkIndex, err := slabPool.locate(chunk)
        if err != nil {
            t.Errorf("unexpected error: %s", err)
        }
        if slab.chunkInfo[chunkIndex].refs != 1 {
            t.Error("chunk refs should be 1")
        }
    }

    testErr(-1)
    testErr(0)
    testErr(4096)

    testOK(1)
    testOK(50)
    testOK(64)
    testOK(512)
    testOK(1024)
}

func TestPut(t *testing.T) {
    slabPool, _ := CreateSlabPool(4096, 64, 1024, 2)

    testErr := func(chunk []byte) {
        err := slabPool.Put(chunk)
        if err == nil {
            t.Error("should return error due to wrong input chunk")
        }
    }

    testOK := func(chunk []byte) {
        slab, chunkIndex, _ := slabPool.locate(chunk)
        chunkFreeBefore := slab.countFree
        chunkRefsBefore := slab.chunkInfo[chunkIndex].refs

        err := slabPool.Put(chunk)
        if err != nil {
            t.Error("should not return error")
        }
        slab, chunkIndex, _ = slabPool.locate(chunk)
        if chunkFreeBefore + 1 != slab.countFree {
            t.Error("count of chunk Free should increase by 1")
        }
        if chunkRefsBefore - 1 != slab.chunkInfo[chunkIndex].refs && 
           chunkRefsBefore != 1 {
            t.Error("reference count should be 0")
        }
    }

    // chunk is nil
    testErr(nil)

    // chunk from golang
    chunk := make([]byte, 100)
    testErr(chunk)

    chunk = make([]byte, 0)
    testErr(chunk)

    // change chunk capacity
    chunk, _ = slabPool.Get(1024)
    chunk = chunk[0:10:10]
    testErr(chunk)

    // chunk from another slabPool
    slabPool2, _ := CreateSlabPool(4096, 64, 1024, 2)
    chunk, _ = slabPool2.Get(128)
    testErr(chunk)

    // chunk from slabPool
    chunk1, _ := slabPool.Get(61)
    chunk2, _ := slabPool.Get(62)
    chunk3, _ := slabPool.Get(63)
    testOK(chunk1)
    testOK(chunk3)
    testOK(chunk2)
}

func TestIncRefAndDecRef(t *testing.T) {
    slabPool, _ := CreateSlabPool(4096, 64, 1024, 2)
    chunk, _ := slabPool.Get(1024)

    slab, chunkIndex, _ := slabPool.locate(chunk)

    // increase reference count
    slabPool.IncRef(chunk)
    if slab.chunkInfo[chunkIndex].refs != 2 {
        t.Errorf("chunk refs should be 2")
    }
    slabPool.IncRef(chunk)
    if slab.chunkInfo[chunkIndex].refs != 3 {
        t.Errorf("chunk refs should be 3")
    }

    // decrease reference count
    slabPool.DecRef(chunk)
    if slab.chunkInfo[chunkIndex].refs != 2 {
        t.Errorf("chunk refs should be 2")
    }
    slabPool.Put(chunk)
    if slab.chunkInfo[chunkIndex].refs != 1 {
        t.Errorf("chunk refs should be 1")
    }
    slabPool.DecRef(chunk)
    if slab.chunkInfo[chunkIndex].refs != 0 {
        t.Errorf("chunk refs should be 0")
    }
}

func TestIncAndDecRef_2(t *testing.T) {
    slabPool, _ := CreateSlabPool(4096, 64, 1024, 2)

    // some abnormal input chunks
    if err := slabPool.IncRef(make([]byte, 2400)); err == nil {
        t.Errorf("should return error")
    }
    if err := slabPool.IncRef(make([]byte, 64)); err == nil {
        t.Errorf("should return error")
    }
    if err := slabPool.DecRef(make([]byte, 64)); err == nil {
        t.Errorf("should return error")
    }
}

func BenchmarkIncAndDecRef(b *testing.B) {
    slabPool, _ := CreateSlabPool(4096, 64, 1024, 2)
    chunk, _ := slabPool.Get(128)

    b.ResetTimer()
    for i:=0; i<b.N; i++ {
        slabPool.IncRef(chunk)
        slabPool.DecRef(chunk)
    }
}

func BenchmarkGetAndPut128(b *testing.B) {
    slabPool, _ := CreateSlabPool(4096, 64, 1024, 2)
    benchmarkGetAndPutSize(b, slabPool, 128)
}

func BenchmarkGetAndPut1024(b *testing.B) {
    slabPool, _ := CreateSlabPool(4096, 64, 1024, 2)
    benchmarkGetAndPutSize(b, slabPool, 128)
}

func BenchmarkGetAndPut16K(b *testing.B) {
    slabPool, _ := CreateSlabPool(4096*1024, 64, 1024*1024, 2)
    benchmarkGetAndPutSize(b, slabPool, 16*1024)
}

func benchmarkGetAndPutSize(b *testing.B, slabPool *SlabPool, size int) {
    b.ResetTimer()
    for i:=0; i<b.N; i++ {
        chunk, _ := slabPool.Get(size)
        slabPool.Put(chunk)
    }
}
