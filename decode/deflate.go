package decode

import(
	"encoding/binary"
)

type Block struct {
	ID int
	Text string
}

type FileData struct {
	Blocks []Block
	XRef XRefBlock
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

func (fileData *FileData) GetMap() map[int]string {
	newMap := make(map[int]string)
	for _, val := range fileData.Blocks {
		newMap[val.ID] = val.Text
	}
	return newMap
}

func BuildXRef(stateWidth, offsetWidth, indexWidth int, data []byte) *XRefBlock {
	totalWidth := stateWidth + offsetWidth + indexWidth
	xRef := new(XRefBlock)
	xRef.Trips = make([]Trip, 0, 32)
	for index := totalWidth - 1; index < len(data); index += totalWidth {
		stateSlice := data[index - totalWidth : index - offsetWidth - indexWidth]
		offsetSlice := data[index - offsetWidth - indexWidth : index - indexWidth]
		indexSlice := data[index - indexWidth : index + 1]
		state, _ := binary.Uvarint(stateSlice)
		offset, _ := binary.Uvarint(offsetSlice)
		xIndex, _ := binary.Uvarint(indexSlice)
		trip := Trip{state, offset, xIndex}
		xRef.Trips = append(xRef.Trips, trip)
	}
	return xRef
}

type XRefBlock struct {
	ID int
	MinIndex, MaxIndex int
	Trips []Trip
}

type Trip struct {
	State, Offset, Index uint64
}

