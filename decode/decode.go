package decode

import(
	"compress/flate"
	"io"
	"io/ioutil"
	"bufio"
	"bytes"
	"regexp"
	"strconv"
	"strings"
	"errors"
	"container/list"
)

const(
	VersionConst = 1.5
)

var NotPDFErr error = errors.New("Does not match PDF specifications")

func getID(line string) (bool, int64) {
	objAddressEx, _ := regexp.Compile("[0-9]+ [0-9]+ obj")
	numberEx, _ := regexp.Compile("[0-9]+")
	line = objAddressEx.FindString(line)
	if line == "" {
		return false, 0
	}
	rawNumber := numberEx.FindString(line)
	address, err := strconv.ParseInt(rawNumber, 10, 64)
	if err != nil {
		return false, 0
	}
	return true, address
}

func getLength(line string, sizeTable map[int64]int64) (bool, int64) {
	lengthEx, _ := regexp.Compile("/Length [0-9]+( [0-9]+ R)?")
	addressEx, _ := regexp.Compile("/Length [0-9]+ [0-9]+ R")
	numberEx, _ := regexp.Compile("[0-9]+")
	line = lengthEx.FindString(line)
	if line == "" {
		return false, 0
	}
	rawNumber := numberEx.FindString(line)
	size, err := strconv.ParseInt(rawNumber, 10, 64)
	if err != nil {
		return false, 0
	}
	if addressEx.MatchString(line) {
		return true, sizeTable[size]
	}
	return true, size
}

func buildSizeTable(toRead io.Reader) (map[int64]int64, error) {
	reader := bufio.NewReader(toRead)
	sizeTable := make(map[int64]int64)
	numberEx, _ := regexp.Compile("[0-9]+")
	var address int64
	for {
		for {
			line, err := reader.ReadString('\n')
			if err == io.EOF {
				goto End
			}
			if err != nil {
				return nil, err
			}
			isAddress := false
			isAddress, address = getID(line)
			if isAddress {
				break
			}
		}
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		if !numberEx.MatchString(line) {
			continue
		}
		rawSize := numberEx.FindString(line)
		size, err := strconv.ParseInt(rawSize, 10, 32)
		if err != nil {
			continue
		}
		sizeTable[address] = size
	}
End:
	return sizeTable, nil
}

func findBlock(toRead io.Reader, sizeTable map[int64]int64) (int, []byte, error) {
	reader := bufio.NewReader(toRead)
	openEx, _ := regexp.Compile("<<")
	closeEx, _ := regexp.Compile(">>")
	headerEx, _ := regexp.Compile("<</Length [0-9]+( [0-9]+ R)?/Filter ?/FlateDecode>>")
	streamEx, _ := regexp.Compile("stream")
	var size int64
	var id int64
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return -1, nil, err
		}
		_, id = getID(line)
		if openEx.MatchString(line) {
			buffer := line
			for !closeEx.MatchString(buffer) {
				line, err = reader.ReadString('\n')
				if err != nil {
					return -1, nil, err
				}
				buffer += line
			}
			buffer = strings.Replace(buffer, "\n", "", -1)
			if !headerEx.MatchString(buffer) {
				continue
			}
			hasSize := false
			hasSize, size = getLength(buffer, sizeTable)
			if !hasSize {
				continue
			}
			if !streamEx.MatchString(buffer) {
				line, err = reader.ReadString('\n')
				if err != nil {
					return -1, nil, err
				}
				if !streamEx.MatchString(line) {
					continue
				}
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
			return -1, nil, err
		}
		bytebuffer[index] = val
	}
	return int(id), bytebuffer, nil
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

func Decode(toRead io.Reader) (*FileData, error) {
	version, err := getVersion(toRead)
	if err != nil {
		return nil, err
	}
	var sizeTable map[int64]int64
	var reader io.Reader
	fileData := New()
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
		id, bytebuffer, err := findBlock(reader, sizeTable)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		bytebuffer, err = ioutil.ReadAll(flate.NewReader(bytes.NewBuffer(bytebuffer)))
		if err != nil {
			return nil, err
		}
		fileData.Append(id, string(bytebuffer))
	}
	return fileData, nil
}

func stitch(readers []io.Reader) io.Reader {
	if len(readers) == 0 {
		return io.LimitReader(nil, 0) //XXX: I want to find something that just returns io.EOF
	}
	multi := readers[0]
	for index := 1; index < len(readers); index++ {
		multi = io.MultiReader(multi, readers[index])
	}
	return multi
}

