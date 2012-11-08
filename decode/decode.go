package decode

import(
	"compress/flate"
	"io"
	"bufio"
	"bytes"
	"regexp"
	"strconv"
)

func Decode(toRead io.Reader) (io.ReadCloser, error) {
	reader := bufio.NewReader(toRead)
	regex, _ := regexp.Compile(".*/Length.*/Filter/FlateDecode.*")
	var size int64
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			return nil, err
		}
		if regex.Match(line) {
			lengthEx, _ := regexp.Compile("[0-9]+")
			referenceEx, _ := regexp.Compile("[0-9]+ [0-9]+ R")
			number := referenceEx.Find(line)
			if number == nil {
				number = lengthEx.Find(line)
				size, err = strconv.ParseInt(string(number), 10, 32)
			} else {
				number = lengthEx.Find(number)
				obj, err := strconv.ParseInt(string(number), 10, 32)
				size = int64(findLength(int(obj), reader))
			}
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

func findLength(obj int, toRead io.Reader) int {
	reader := bufio.NewReader(toRead)
	numString := strconv.Format(int64(obj), 10)
	regex, _ := regexp.Compile(numString + " [0-9]+ obj")
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			return nil, err
		}
		if regex.Match(line) {
			break
		}
	}
	regex, _ = regexp.Compile("[0-9]+")
	line, err := reader.ReadBytes('\n')
	number := regex.Find(line)
	toReturn, err := strconv.ParseInt(string(number), 10, 32)
	return int(toReturn)
}

