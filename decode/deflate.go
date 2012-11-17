package decode

import(
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

func Uint64(stream []byte) uint64 {
	var out uint64
	for _, val := range stream {
		out *= 256
		out += uint64(val)
	}
	return out
}

func BuildXRef(stateWidth, offsetWidth, indexWidth int, data []byte) *XRefBlock {
	totalWidth := stateWidth + offsetWidth + indexWidth
	xRef := new(XRefBlock)
	xRef.Trips = make([]Trip, 0, 32)
	for index := 0; index + totalWidth < len(data); index += totalWidth {
		stateSlice := data[index : index + stateWidth]
		offsetSlice := data[index + stateWidth : index + stateWidth + offsetWidth]
		indexSlice := data[index + stateWidth + offsetWidth : index + totalWidth]
		state := Uint64(stateSlice)
		offset := Uint64(offsetSlice)
		xIndex := Uint64(indexSlice)
		trip := Trip{state, offset, xIndex}
		xRef.Trips = append(xRef.Trips, trip)
	}
	return xRef
}

type XRefBlock struct {
	ID int
	MinIndex, MaxIndex int
	Trips []Trip
	ExtendedTable map[int64][]byte
}

func (xRef *XRefBlock) BuildXTable(file io.ReadSeeker) {
	xRef.ExtendedTable = make(map[int64]byte)
	for _, val := range xRef.Trips {
		if val.State == 2 {
			_, inMap = xRef.ExtendedTable[val.Offset]
			if !inMap {
				data, err := findBlock(file, val.Offset, xRef)
				data, err = Uncompress(data)
				xRef.ExtendedTable[val.Offset] = data
			}
		}
	}
}

func (xRef *XRefBlock) GetOffset(id int64) {
	return xRef.Trips[id].Offset
}

type Trip struct {
	State, Offset, Index uint64
}

type RootBlock struct {
	TreeRoot, Outline int
	Lang string
	Metadata, Pages int
	PageLabels []string
}

