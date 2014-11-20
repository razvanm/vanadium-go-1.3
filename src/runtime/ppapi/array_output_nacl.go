package ppapi

type arrayOutputBuffer struct {
	buffer      []byte
	count, size uint32
}

func get_array_output_buffer(alloc *arrayOutputBuffer, count uint32, size uint32) *byte {
	alloc.count = count
	alloc.size = size
	if count == 0 || size == 0 {
		return nil
	}
	alloc.buffer = make([]byte, count*size)
	return &alloc.buffer[0]
}
