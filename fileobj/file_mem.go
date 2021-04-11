package fileobj

type MemFileObj struct {
	data       []byte
	fileSize   int64
	blockSize  int64
	blockCount int64
}

func NewMemFileObj(fileSize, blockSize int64) (obj *MemFileObj, err error) {
	obj = &MemFileObj{
		data:      make([]byte, fileSize),
		fileSize:  fileSize,
		blockSize: blockSize,
	}
	obj.updateBlockCount()
	return obj, nil
}

func (obj *MemFileObj) Close() error {
	return nil
}

func (obj *MemFileObj) FileSize() int64 {
	return obj.fileSize
}

func (obj *MemFileObj) BlockSize() int64 {
	return obj.blockSize
}

func (obj *MemFileObj) BlockCount() int64 {
	return obj.blockCount
}

func (obj *MemFileObj) updateMeta() (err error) {
	obj.fileSize = int64(len(obj.data))
	obj.updateBlockCount()
	return
}

func (obj *MemFileObj) updateBlockCount() {
	count, rem := obj.fileSize/obj.blockSize, obj.fileSize%obj.blockSize
	if rem > 0 {
		count += 1
	}
	obj.blockCount = count
}

func (obj *MemFileObj) GetBlock(blk int64) (data []byte, err error) {
	if obj.blockCount <= blk {
		return nil, ErrOutOfBlockIndex
	}
	data = make([]byte, obj.blockSize)
	copy(data, obj.data[obj.blockSize*blk:])
	return data, nil
}

func (obj *MemFileObj) SetBlock(blk int64, data []byte) (err error) {
	if obj.blockCount <= blk {
		return ErrOutOfBlockIndex
	}
	if int64(len(data)) != obj.blockSize {
		return ErrWrongBlockSize
	}
	copy(obj.data[obj.blockSize*blk:], data)
	if err = obj.updateMeta(); err != nil {
		return err
	}
	return
}
