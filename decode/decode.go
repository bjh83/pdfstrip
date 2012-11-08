package decode

import(
	"compress/flate"
	"container/list"
	"os"
	"io"
	"bufio"
	"bytes"
	"regexp"
	"fmt"
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
			fmt.Println("broke")
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
			fmt.Println("broke")
			break
		}
		for _, val := range line {
			byteList.PushBack(val)
		}
	}
	toPrepend := []byte{31, 139, 8, 0, 0, 0, 0, 0, 0, 1, 67, 68, 69, 70, 71}
	bytebuffer := make([]byte, byteList.Len() - 2 + len(toPrepend))
	var index int
	for index, val := range toPrepend {
		bytebuffer[index] = val
	}
	byteList.Remove(byteList.Front())
	byteList.Remove(byteList.Front())
	for element := byteList.Front(); element != nil; element = element.Next() {
		bytebuffer[index] = element.Value.(byte)
		index++
	}
	fmt.Println(len(bytebuffer))
	buffer := bytes.NewBuffer(bytebuffer)
	return flate.NewReader(buffer), nil
}

