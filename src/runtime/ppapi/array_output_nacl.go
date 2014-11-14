package ppapi

import "log"

type arrayOutputBuffer struct {
	buffer      []byte
	count, size uint32
}

func get_array_output_buffer(alloc *arrayOutputBuffer, count uint32, size uint32) *byte {
	log.Printf("AAAAAAAA1a: %p:%v, count=%d, size=%d", alloc, alloc, count, size)
	alloc.count = count
	alloc.size = size
	if count == 0 || size == 0 {
		return nil
	}
	alloc.buffer = make([]byte, count*size)
	log.Printf("AAAAAAAA1b: %v", alloc)
	return &alloc.buffer[0]
}
