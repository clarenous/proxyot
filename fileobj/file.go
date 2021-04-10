package fileobj

import (
	"errors"
	"os"
)

const (
	KiB = 1024
	MiB = 1024 * KiB
	GiB = 1024 * MiB
)

var (
	ErrOutOfBlockIndex = errors.New("out of data block index")
	ErrWrongBlockSize  = errors.New("wrong data block size")
)

type FileObj struct {
	f          *os.File
	fileSize   int64
	blockSize  int64
	blockCount int64
}

// TODO: support readonly/readwrite mode
func NewFileObj(filename string, blockSize int64) (obj *FileObj, err error) {
	var f *os.File
	if f, err = os.Open(filename); err != nil {
		return nil, err
	}
	var fi os.FileInfo
	if fi, err = f.Stat(); err != nil {
		f.Close()
		return nil, err
	}
	obj = &FileObj{
		f:         f,
		fileSize:  fi.Size(),
		blockSize: blockSize,
	}
	obj.updateBlockCount()
	return obj, nil
}

func (obj *FileObj) FileSize() int64 {
	return obj.fileSize
}

func (obj *FileObj) BlockSize() int64 {
	return obj.blockSize
}

func (obj *FileObj) BlockCount() int64 {
	return obj.blockCount
}

func (obj *FileObj) updateMeta() (err error) {
	var fi os.FileInfo
	if fi, err = obj.f.Stat(); err != nil {
		return
	}
	obj.fileSize = fi.Size()
	obj.updateBlockCount()
	return
}

func (obj *FileObj) updateBlockCount() {
	count, rem := obj.fileSize/obj.blockSize, obj.fileSize%obj.blockSize
	if rem > 0 {
		count += 1
	}
	obj.blockCount = count
}

func (obj *FileObj) GetBlock(blk int64) (data []byte, err error) {
	if obj.blockCount <= blk {
		return nil, ErrOutOfBlockIndex
	}
	data = make([]byte, obj.blockSize)
	if _, err = obj.f.ReadAt(data, obj.blockSize*blk); err != nil {
		return nil, err
	}
	return data, nil
}

func (obj *FileObj) SetBlock(blk int64, data []byte) (err error) {
	if obj.blockCount <= blk {
		return ErrOutOfBlockIndex
	}
	if int64(len(data)) != obj.blockSize {
		return ErrWrongBlockSize
	}
	if _, err = obj.f.WriteAt(data, obj.blockSize*blk); err != nil {
		return err
	}
	if err = obj.updateMeta(); err != nil {
		return err
	}
	return
}
