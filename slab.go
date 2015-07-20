/* slab.go - Slab implementation */
/*
modification history
--------------------
2014/12/2, by Sijie Yang, create
*/
/*
DESCRIPTION
*/
package slab_pool

import (
    "encoding/binary"
    "unsafe"
)

const SLAB_FOOTER_LEN int = 16

type Slab struct {
    slabSize   int         // slab size
    chunkSize  int         // chunk size
    slabMagic  uint64      // magic number for slab footer

    memory     []byte      // slab memory area
    chunkInfo  []ChunkInfo // chunk info
    chunkFree  int         // head of chunk free list

    countChunk int         // count of chunks in this slab
    countFree  int         // count of free chunks in this slab

    /* management info in its slabClass */
    slabClass  *SlabClass  // link to its slabClass
    index      int         // slab index of slabClass.slabs
    whichList  int         // in which slablist (SLAB_FREE/SLAB_USE/SLAB_FULL)
    prev       int         // prev node in slablist
    next       int         // next node in slablist
}

func NewSlab(sc *SlabClass, slabSize int, chunkSize int, slabMagic uint64) *Slab {
    s := new(Slab)
    s.slabClass = sc
    s.slabSize = slabSize
    s.chunkSize = chunkSize
    s.slabMagic = slabMagic

    // TODO(yangsijie): alloc slab memory from pool
    s.memory = make([]byte, slabSize+SLAB_FOOTER_LEN)

    // initial chunk info
    s.countChunk = slabSize / chunkSize
    s.countFree = s.countChunk
    s.chunkInfo = make([]ChunkInfo, s.countChunk)
    s.chunkFree = 0
    s.initChunkInfo()

    // initial link info
    s.index = -1
    s.whichList = -1
    s.prev = -1
    s.next = -1

    s.initFooter()
    return s
}

// initial chunk info
func (s *Slab) initChunkInfo() {
    var i = 0
    for ; i < len(s.chunkInfo)-1; i++ {
        s.chunkInfo[i].next = i + 1
    }
    s.chunkInfo[i].next = -1
}

// initial footer info
func (s *Slab) initFooter() {
    footer := s.memory[s.slabSize:]
    slabPtr := uintptr(unsafe.Pointer(s))
    binary.BigEndian.PutUint64(footer[0:8], s.slabMagic)      // slab magic
    binary.BigEndian.PutUint64(footer[8:16], uint64(slabPtr)) // slab pointer
}

// allocate chunk
func (s *Slab) chunkAlloc() []byte {
    if s.countFree <= 0 {
        return nil
    }

    // remove head chunk from free list
    head := s.chunkFree
    s.chunkFree = s.chunkInfo[head].next
    s.chunkInfo[head].next = -1
    s.chunkInfo[head].refs = 1
    s.countFree -= 1

    // return chunk
    return s.memory[s.chunkSize*head : s.chunkSize*(head+1)]
}

// increase refs for chunk
func (s *Slab) chunkIncRef(index int) {
    s.chunkInfo[index].incRef()
}

// decrease refs for chunk
func (s *Slab) chunkDecRef(index int) {
    // add chunk to free list
    if s.chunkInfo[index].decRef() == 0 {
        s.chunkInfo[index].next = s.chunkFree
        s.chunkFree = index
    }
    s.countFree += 1
}

// return slab status
func (s *Slab) status() int {
    if s.countFree == s.countChunk {
        return SLAB_FREE
    } else if s.countFree == 0 {
        return SLAB_FULL
    } else {
        return SLAB_USE
    }
}
