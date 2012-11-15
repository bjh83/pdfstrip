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
	return &FileData{}
}

func (fileData *FileData) Append(id int, text string) {
	if fileData.Blocks == nil {
		fileData.Blocks = []Block{Block{id, text}}
	} else {
		fileData.Blocks = append(fileData.Blocks, Block{id, text})
	}
}

