/* slab_pool.go - SlabPool implementation */
/*
modification history
--------------------
2014/12/2, by Sijie Yang, create
*/
/*
DESCRIPTION

Usage:
    // Create slab pool
    slabPool, err := CreateSlabPool(4096, 128, 1024, 2)

    // Allocate chunk
    chunk, err := slabPool.Get(500)

    // Release chunk
    err := slabPool.Put(chunk)

    // Chunk reference operations
    chunk2, err := slabPool.Get(500)
    slabPool.IncRef(chunk2)
    slabPool.DecRef(chunk2)

Note:
    Must Not append() on chunk allocated.
*/
package slab_pool

import (
    "encoding/binary"
    "fmt"
    "math/rand"
    "sort"
    "unsafe"
)

type SlabPool struct {
    slabClasses  []SlabClass // SlabClasses with different chunk size

    slabSize     int         // slab size (bytes)
    chunkSizeMax int         // max chunk size (bytes)
    chunkSizeMin int         // min chunk size (bytes)
    factor       float64     // growth factor for chunk size

    slabMagic    uint64      // magic number for slab
}

/* CreateSlabPool - create slab pool
 * 
 * Params:
 *     - slabSize    : size of slab (bytes)
 *     - chunkSizeMin: min chunk size (bytes)
 *     - chunkSizeMax: max chunk size (bytes)
 *     - factor      : growth factor for chunk size 
 * 
 * Return:
 *     - slabPool    : slab pool
 *     - error       : nil if success, error if failure
 */
func CreateSlabPool(slabSize int, chunkSizeMin int, chunkSizeMax int, factor float64) (
    *SlabPool, error) {
    if err := validateParams(slabSize, chunkSizeMin, chunkSizeMax, factor); err != nil {
        return nil, fmt.Errorf("wrong params: %s", err)
    }

    sp := new(SlabPool)
    sp.slabSize = slabSize
    sp.chunkSizeMax = chunkSizeMax
    sp.chunkSizeMin = chunkSizeMin
    sp.factor = factor
    sp.slabMagic = uint64(rand.Int63())
    sp.initSlabClass()

    return sp, nil
}

// validate parameters for init slabpool
func validateParams(slabSize int, chunkSizeMin int, chunkSizeMax int, factor float64) error {
    if chunkSizeMin <= 0 || chunkSizeMin > chunkSizeMax {
        return fmt.Errorf("wrong chunkSizeMin/chunkSizeMax")
    }
    if slabSize < chunkSizeMax {
        return fmt.Errorf("slabSize should be no less than chunkSizeMax")
    }
    if factor <= 1 {
        return fmt.Errorf("factor should be greater than 1")
    }
    if int(float64(chunkSizeMin)*factor) == chunkSizeMin {
        return fmt.Errorf("wrong chunkSizeMin/factor")
    }
    return nil
}

// initial slabclasses
func (sp *SlabPool) initSlabClass() {
    sp.slabClasses = make([]SlabClass, 0)

    chunkSize := sp.chunkSizeMin
    for chunkSize <= sp.chunkSizeMax {
        slabClass := NewSlabClass(sp.slabSize, chunkSize, sp.slabMagic)
        sp.slabClasses = append(sp.slabClasses, *slabClass)

        chunkSize = int((float64(chunkSize) * sp.factor))
    }
}

/* Get - allocate a chunk with length 'size'
 *
 * Params:
 *     - size: chunk size
 *
 * Return:
 *     - chunk: chunk allocated
 *     - err  : error
 *
 * Note:
 *     Must Not apppend() on return chunk
 */
func (sp *SlabPool) Get(size int) ([]byte, error) {
    if size > sp.chunkSizeMax || size <= 0 {
        return nil, fmt.Errorf("illegal chunk size: %s", size)
    }

    // find slab class by chunk size
    slabClass := sp.slabClassFor(size)

    // get free chunk from slab class
    chunk, err := slabClass.chunkAlloc()
    if err != nil {
        return nil, fmt.Errorf("Get(): %s", err.Error())
    }
    return chunk[:size], nil
}

/* Put - release chunk to slab pool
 *
 * Params:
 *     - chunk: chunk to release
 *
 * Return:
 *     - err: error
 */
func (sp *SlabPool) Put(chunk []byte) error {
    return sp.DecRef(chunk)
}

/* IncRef - increase reference for chunk
 *
 * Params:
 *     chunk: chunk allocated
 *
 * Return:
 *     err: error
 */
func (sp *SlabPool) IncRef(chunk []byte) error {
    if err := sp.validateChunk(chunk); err != nil {
        return err
    }

    // find slab for this chunk
    slab, chunkIndex, err := sp.locate(chunk)
    if err != nil {
        return err
    }

    // increase reference count for chunk
    slabClass := slab.slabClass
    slabClass.chunkIncRef(slab, chunkIndex)
    return nil
}

/* DecRef - decrease reference for chunk
 *
 * Params:
 *     chunk: chunk allocated
 *
 * Return:
 *     err: error
 */
func (sp *SlabPool) DecRef(chunk []byte) error {
    if err := sp.validateChunk(chunk); err != nil {
        return err
    }

    // find slab for this chunk
    slab, chunkIndex, err := sp.locate(chunk)
    if err != nil {
        return err
    }

    // decrease reference count for chunk
    slabClass := slab.slabClass
    slabClass.chunkDecRef(slab, chunkIndex)
    return nil
}

// validate input chunk
func (sp *SlabPool) validateChunk(chunk []byte) error {
    // check chunk not nil
    if chunk == nil {
        return fmt.Errorf("chunk is nil")
    }
    // check chunk size
    if len(chunk) <= 0 || len(chunk) > sp.chunkSizeMax {
        return fmt.Errorf("chunk size should be no greater than %s", sp.chunkSizeMax)
    }
    // check chunk capacity (must not be changed)
    if cap(chunk) <= SLAB_FOOTER_LEN {
        return fmt.Errorf("chunk capicity should not be changed")
    }
    return nil
}

// find slabClass with matched chunksize
func (sp *SlabPool) slabClassFor(size int) *SlabClass {
    i := sort.Search(len(sp.slabClasses),
        func(i int) bool {
            return size <= sp.slabClasses[i].chunkSize
        })
    return &(sp.slabClasses[i])
}

// find slab and chunkIndex for input chunk
func (sp *SlabPool) locate(chunk []byte) (*Slab, int, error) {
    // locate footer area
    restArea := chunk[:cap(chunk)]
    footer := restArea[len(restArea)-SLAB_FOOTER_LEN:]

    // read slab pointer
    slab, err := sp.getSlabPtr(footer)
    if err != nil {
        return nil, -1, err
    }

    // return slab and chunk index
    chunkIndex := (slab.slabSize + SLAB_FOOTER_LEN - cap(chunk)) / slab.chunkSize
    return slab, chunkIndex, nil
}

// get slab pointer from footer area
func (sp *SlabPool) getSlabPtr(footer []byte) (*Slab, error) {
    slabMagic := binary.BigEndian.Uint64(footer[0:8])
    slabPtr := binary.BigEndian.Uint64(footer[8:16])
    if slabMagic != sp.slabMagic { // validate slabPtr by slabMagic
        return nil, fmt.Errorf("magic number not matched, a chunk not allocted from this pool?")
    }
    return (*Slab)(unsafe.Pointer(uintptr(slabPtr))), nil
}
