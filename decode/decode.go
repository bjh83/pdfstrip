package decode

import(
	"compress/flate"
	"io"
	"io/ioutil"
	"bufio"
	"bytes"
	"regexp"
	"strconv"
	"errors"
	"container/list"
	"fmt"
)

const(
	VersionConst = 1.6
)

var NotPDFErr error = errors.New("Does not match PDF specifications")

func buildSizeTable(toRead io.Reader) (map[int64]int64, error) {
	reader := bufio.NewReader(toRead)
	sizeTable := make(map[int64]int64)
	objAddressEx, _ := regexp.Compile("[0-9]+ [0-9]+ obj\n")
	isNumberEx, _ := regexp.Compile("[0-9]+\n")
	numberEx, _ := regexp.Compile("[0-9]+")
	for {
		var address int64
		for {
			line, err := reader.ReadString('\n')
			if err == io.EOF {
				goto End
			}
			if err != nil {
				return nil, err
			}
			if objAddressEx.MatchString(line) {
				rawAddress := numberEx.FindString(line)
				address, err = strconv.ParseInt(rawAddress, 10, 32)
				if err != nil {
					return nil, err
				}
				break
			}
		}
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		if !isNumberEx.MatchString(line) {
			continue
		}
		rawSize := numberEx.FindString(line)
		size, err := strconv.ParseInt(rawSize, 10, 32)
		if err != nil {
			return nil, err
		}
		sizeTable[address] = size
	}
End:
	return sizeTable, nil
}

func findBlock(toRead io.Reader, sizeTable map[int64]int64) ([]byte, error) {
	reader := bufio.NewReader(toRead)
	lenStmtEx, _ := regexp.Compile("^(<<)?/Length.*")
	filterEx, _ := regexp.Compile(".*/Filter ?/FlateDecode.*")
	numberEx, _ := regexp.Compile("[0-9]+")
	old := sizeTable != nil
	var size int64
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		if lenStmtEx.MatchString(line) {
			rawSize := numberEx.FindString(line)
			fmt.Print(line)
			size, err = strconv.ParseInt(rawSize, 10, 32)
			if err != nil {
				return nil, err
			}
			if old {
				size = sizeTable[size]
			}
			if size < 3 {
				continue
			}
			if !filterEx.MatchString(line) {
				line, err := reader.ReadString('\n')
				if err != nil {
					return nil, err
				}
				if !filterEx.MatchString(line) {
					continue
				}
				fmt.Print(line)
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
	reader.ReadByte()
	reader.ReadByte()
	for ;index < len(bytebuffer); index++ {
		val, err := reader.ReadByte()
		if err != nil {
			return nil, err
		}
		bytebuffer[index] = val
	}
	return bytebuffer, nil
}

func getVersion(toRead io.Reader) (float32, error) {
	versionInfo := make([]byte, 9)
	_, err := toRead.Read(versionInfo)
	if err != nil {
		return 0, err
	}
	regex, _ := regexp.Compile("%PDF-[0-9]+\\.[0-9]+")
	if !regex.Match(versionInfo) {
		return 0, NotPDFErr
	}
	regex, _ = regexp.Compile("[0-9]+\\.[0-9]+")
	rawVersion := regex.Find(versionInfo)
	version, err := strconv.ParseFloat(string(rawVersion), 32)
	if err != nil {
		return 0, err
	}
	return float32(version), nil
}

func Decode(toRead io.Reader) (io.Reader, error) {
	version, err := getVersion(toRead)
	if err != nil {
		return nil, err
	}
	var sizeTable map[int64]int64
	var reader io.Reader
	if version < VersionConst {
		bytebuffer, err := ioutil.ReadAll(toRead)
		//I know this is not great but the alternatives are worse
		// and I want this code to be generic
		buffer := bytes.NewBuffer(bytebuffer)
		sizeTable, err = buildSizeTable(buffer)
		if err != nil {
			return nil, err
		}
		reader = bufio.NewReader(bytes.NewBuffer(bytebuffer))
	} else {
		reader = bufio.NewReader(toRead)
	}
	readerList := list.New()
	readerList.Init()
	for {
		bytebuffer, err := findBlock(reader, sizeTable)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		readerList.PushBack(flate.NewReader(bytes.NewBuffer(bytebuffer)))
	}
	readerArray := make([]io.Reader, readerList.Len())
	for element, index := readerList.Front(), 0; element != nil; element, index = element.Next(), index + 1 {
		readerArray[index] = element.Value.(io.Reader)
	}
	return stitch(readerArray), nil
}

func stitch(readers []io.Reader) io.Reader {
	multi := readers[0]
	for index := 1; index < len(readers); index++ {
		multi = io.MultiReader(multi, readers[index])
	}
	return multi
}
