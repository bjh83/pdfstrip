package decode

import(
	"compress/flate"
	"container/list"
	"os"
	"io"
	"bufio"
	"bytes"
	"regexp"
)

func Decode(file *os.File) (io.ReadCloser, error) {
	reader := bufio.NewReader(file)
	regex, _ := regexp.Compile(".*/Filter/FlateDecode.*")
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			return nil, err
		}
		if regex.Match(line) {
			break
		}
	}
	regex, _ = regexp.Compile("endstream")
	byteList := list.New()
	byteList.Init()
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			return nil, err
		}
		if regex.Match(line) {
			break
		}
		for _, val := range line {
			byteList.PushBack(val)
		}
	}
	bytebuffer := make([]byte, byteList.Len())
	index := 0
	for element := byteList.Front(); element != nil; element = element.Next() {
		bytebuffer[index] = element.Value.(byte)
		index++
	}
	buffer := bytes.NewBuffer(bytebuffer)
	return flate.NewReader(buffer), nil
}

