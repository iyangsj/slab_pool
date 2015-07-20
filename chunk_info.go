/* chunk_info.go - chunk info */
/*
modification history
--------------------
2014/12/2, by Sijie Yang, create
*/
/*
DESCRIPTION
    For more information, see BFE-Cache 2-004
*/
package slab_pool

import (
    "fmt"
)

type ChunkInfo struct {
    refs int // reference count
    next int // next node in the chunk free list
}

// increase reference
func (c *ChunkInfo) incRef() int {
    c.refs++
    if c.refs <= 1 {
        panic(fmt.Sprintf("incRef(): unexpected reference count %d: %#v", c.refs, c))
    }
    return c.refs
}

// decrease reference
func (c *ChunkInfo) decRef() int {
    c.refs--
    if c.refs < 0 {
        panic(fmt.Sprintf("decRef(): unexpected reference count %d: %#v", c.refs, c))
    }
    return c.refs
}
