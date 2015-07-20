# slab-pool - yet another slab allocator

# Usage
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

# Limitation
 * Must Not append() on chunk allocated.

# License
Apache License Version 2.0
