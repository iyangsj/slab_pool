/* slab_class.go - SlabClass implementation */
/*
modification history
--------------------
2014/12/2, by Sijie Yang, create
*/
/*
DESCRIPTION
*/
package slab_pool

const (
    SLAB_FREE = 0 // slab with no chunk allocated
    SLAB_USE  = 1 // slab with some chunk allocated
    SLAB_FULL = 2 // slab with all chunk allocated
)

type SlabClass struct {
    slabSize     int      // slab size
    chunkSize    int      // chunk size
    slabMagic    uint64   // magic number for slab

    slabs        []*Slab  // all slabs
    slabLists[3] int      // head of slab lists(SLAB_FREE/SLAB_USE/SLAB_FULL)
}

func NewSlabClass(slabSize int, chunkSize int, slabMagic uint64) *SlabClass {
    sc := new(SlabClass)
    sc.slabSize = slabSize
    sc.chunkSize = chunkSize
    sc.slabMagic = slabMagic

    /* slab lists */
    sc.slabs = make([]*Slab, 0, 10)
    sc.slabLists[SLAB_FREE] = -1
    sc.slabLists[SLAB_USE] = -1
    sc.slabLists[SLAB_FULL] = -1

    return sc
}

// allocate slab
func (sc *SlabClass) slabAlloc() *Slab {
    slab := NewSlab(sc, sc.slabSize, sc.chunkSize, sc.slabMagic)
    sc.slabs = append(sc.slabs, slab)
    slab.index = len(sc.slabs) - 1
    return slab
}

// allocate chunk
func (sc *SlabClass) chunkAlloc() ([]byte, error) {
    // 1. try to alloc chunk from slabsUse list
    if !sc.listEmpty(SLAB_USE) {
        head := sc.slabLists[SLAB_USE]
        slab := sc.slabs[head]
        chunk := slab.chunkAlloc()

        if slab.status() == SLAB_FULL {
            // remove from slabsUse and add to slabFull
            sc.listRemove(SLAB_USE, head)
            sc.listAdd(SLAB_FULL, head)
        }
        return chunk, nil
    }

    // 2. try to alloc chunk from slabsFree list
    if !sc.listEmpty(SLAB_FREE) {
        head := sc.slabLists[SLAB_FREE]
        slab := sc.slabs[head]
        chunk := slab.chunkAlloc()

        // remove from slabsFree list
        sc.listRemove(SLAB_FREE, head)
        if slab.status() == SLAB_FULL {
            // add to slabsFull list
            sc.listAdd(SLAB_FULL, head)
        } else {
            // add to slabsUse list
            sc.listAdd(SLAB_USE, head)
        }
        return chunk, nil
    }

    // 3. try to alloc new slab
    slab := sc.slabAlloc()
    chunk := slab.chunkAlloc()
    if slab.status() == SLAB_FULL {
        sc.listAdd(SLAB_FULL, slab.index)
    } else {
        sc.listAdd(SLAB_USE, slab.index)
    }
    return chunk, nil
}

// increase refs for chunk
func (sc *SlabClass) chunkIncRef(slab *Slab, chunkIndex int) {
    slab.chunkIncRef(chunkIndex)
}

// decrease refs for chunk
func (sc *SlabClass) chunkDecRef(slab *Slab, chunkIndex int) {
    statusBefore := slab.status()

    // decrease refs for chunk
    slab.chunkDecRef(chunkIndex)

    // move slab to new slablist
    statusAfter := slab.status()
    if statusBefore != statusAfter {
         sc.listRemove(statusBefore, slab.index)
         sc.listAdd(statusAfter, slab.index)
    }
}

// remove node from list 'whichList'
func (sc *SlabClass) listRemove(whichList int, node int) {
    head := sc.slabLists[whichList]
    if node == head { // remove head from list
        next := sc.slabs[node].next
        if next >= 0 { // if next is valid node
            sc.slabs[next].prev = -1
        }
        sc.slabLists[whichList] = next
    } else {
        prev := sc.slabs[node].prev
        next := sc.slabs[node].next
        sc.slabs[prev].next = next
        if next >= 0 { // if next is valid node
            sc.slabs[next].prev = prev
        }
    }

    // clear node state
    sc.slabs[node].whichList = -1
    sc.slabs[node].prev = -1
    sc.slabs[node].next = -1
}

// add node to list 'whichList'
func (sc *SlabClass) listAdd(whichList int, node int) {
    head := sc.slabLists[whichList]
    // add node to head of list
    sc.slabs[node].next = head
    sc.slabs[node].prev = -1
    sc.slabs[node].whichList = whichList
    sc.slabLists[whichList] = node
    if head >= 0 { // if head is valid node
        sc.slabs[head].prev = node
    }
}

// check list 'whichlist' is empty or not
func (sc *SlabClass) listEmpty(whichList int) bool {
    return sc.slabLists[whichList] < 0
}
