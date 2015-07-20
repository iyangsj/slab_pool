/* chunk_info_test.go - unit test for chunk_info.go */
/*
modification history
--------------------
2014/12/3, by Sijie Yang, create
*/
/*
DESCRIPTION
    For more information, see BFE-Cache 2-004
*/
package slab_pool

import (
    "testing"
)

func TestChunkInfo(t *testing.T) {
    chunkInfo := new(ChunkInfo)
    chunkInfo.refs = 1

    // test addRef()
    chunkInfo.incRef()
    if chunkInfo.refs != 2 {
        t.Errorf("refs should equal 1")
    }
    chunkInfo.decRef()
    if chunkInfo.refs != 1 {
        t.Errorf("refs should equal 2")
    }

    // test decRef()
    chunkInfo.decRef()
    if chunkInfo.refs != 0 {
        t.Errorf("refs should equal 1")
    }
}

func TestIncRefPanic(t *testing.T) {
    noPanic := false
    defer func() {
        if recover() != nil && noPanic {
            t.Errorf("should panic when refs is wrong")
        }
    }()

    chunkInfo := new(ChunkInfo)
    chunkInfo.incRef()
    noPanic = true
}

func TestDecRefPanic(t *testing.T) {
    noPanic := false
    defer func() {
        if recover() != nil && noPanic {
            t.Errorf("should panic when refs is wrong")
        }
    }()

    chunkInfo := new(ChunkInfo)
    chunkInfo.decRef()
    noPanic = true
}
