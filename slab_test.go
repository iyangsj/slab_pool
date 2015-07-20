/* slab_test.go - unit test for slab.go */
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

import "testing"

func TestChunkAlloc(t *testing.T) {
    slab := NewSlab(nil, 4096, 2048, 201412)

    // allocate chunk
    chunk := slab.chunkAlloc()
    if chunk == nil || len(chunk) != 2048 {
        t.Errorf("shuold return valid chunk with size 2048")
    }
    chunk = slab.chunkAlloc()

    // no free chunks any more
    chunk = slab.chunkAlloc()
    if chunk != nil {
        t.Errorf("there is no more chunk, should return nil")
    }
}
