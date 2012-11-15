package decode

import()

type Block struct {
	ID int
	Text string
}

type FileData struct {
	Blocks []Block
}

func New() *FileData {
	return &FileData{make([]Block, 16)}
}

func (fileData *FileData) Append(id int, text string) {
	fileData.Blocks = append(fileData.Blocks, Block{id, text})
}

