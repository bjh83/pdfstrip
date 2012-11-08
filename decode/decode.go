package decode

import(
	"compress/flate"
	"os"
	"io"
	"bufio"
	"bytes"
	"regexp"
	"strconv"
)

func Decode(file *os.File) (io.ReadCloser, error) {
	reader := bufio.NewReader(file)
	regex, _ := regexp.Compile(".*/Length.*/Filter/FlateDecode.*")
	var size int64
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			return nil, err
		}
		if regex.Match(line) {
			regex, _ = regexp.Compile("[0-9]+")
			number := regex.Find(line)
			size, err = strconv.ParseInt(string(number), 10, 32)
			if err != nil {
				return nil, err
			}
			break
		}
	}
	toPrepend := []byte{31, 139, 8, 0, 0, 0, 0, 0, 0, 1, 67, 68, 69, 70, 71}
	bytebuffer := make([]byte, int(size) - 2 + len(toPrepend))
	var index int
	for index, val := range toPrepend {
		bytebuffer[index] = val
	}
	_, err := reader.ReadByte()
	_, err = reader.ReadByte()
	if err != nil {
		return nil, err
	}
	for ;index < len(bytebuffer); index++ {
		val, err := reader.ReadByte()
		if err != nil {
			return nil, err
		}
		bytebuffer[index] = val
	}
	buffer := bytes.NewBuffer(bytebuffer)
	return flate.NewReader(buffer), nil
}

