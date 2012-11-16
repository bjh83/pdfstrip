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
)

const(
	VersionConst = 1.5
)

var NotPDFErr error = errors.New("Does not match PDF specifications")

var Dictionary []byte = []byte{31, 139, 8, 0, 0, 0, 0, 0, 0, 1, 67, 68, 69, 70, 71}

func GetID(line string) (bool, int64) {
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

func GetLength(line string, sizeTable map[int64]int64) (bool, int64) {
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
			isAddress, address = GetID(line)
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

func findBlock(toRead io.Reader, sizeTable map[int64]int64, headerEx *regexp.Regexp) (int, []byte, string, error) {
	reader := bufio.NewReader(toRead)
	openEx, _ := regexp.Compile("<<")
	closeEx, _ := regexp.Compile(">>")
	streamEx, _ := regexp.Compile("stream")
	var size int64
	var id int64
	var header string
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return -1, nil, "", err
		}
		hasId, newId := GetID(line)
		if hasId {
			id = newId
		}
		if openEx.MatchString(line) {
			buffer := line
			for !closeEx.MatchString(buffer) {
				line, err = reader.ReadString('\n')
				if err != nil {
					return -1, nil, "", err
				}
				buffer += line
			}
			buffer = strings.Replace(buffer, "\n", "", -1)
			if !headerEx.MatchString(buffer) {
				continue
			}
			hasSize := false
			hasSize, size = GetLength(buffer, sizeTable)
			if !hasSize {
				continue
			}
			if !streamEx.MatchString(buffer) {
				line, err = reader.ReadString('\n')
				if err != nil {
					return -1, nil, "", err
				}
				if !streamEx.MatchString(line) {
					continue
				}
			}
			break
			header = buffer
		}
	}
	bytebuffer := make([]byte, int(size) - 2 + len(Dictionary))
	var index int
	for index, val := range Dictionary {
		bytebuffer[index] = val
	}
	reader.ReadByte()
	reader.ReadByte()
	for ;index < len(bytebuffer); index++ {
		val, err := reader.ReadByte()
		if err != nil {
			return -1, nil, "",  err
		}
		bytebuffer[index] = val
	}
	return int(id), bytebuffer, header, nil
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
	headerEx, _ := regexp.Compile("<</Length [0-9]+( [0-9]+ R)?/Filter ?/FlateDecode>>")
	for {
		id, bytebuffer, _, err := findBlock(reader, sizeTable, headerEx)
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

func uncompress(stream []byte) ([]byte, error) {
	uncompressed, err := ioutil.ReadAll(flate.NewReader(bytes.NewBuffer(stream)))
	if err != nil {
		return nil, err
	}
	return uncompressed, nil
}

func GetXRef(toRead io.Reader) (*XRefBlock, error) {
	headerEx, _ := regexp.Compile("<</Type/XRef/W\\[[0-9] [0-9] [0-9]\\]/Root [0-9]+ [0-9]+ R/Index\\[[0-9]+ [0-9]+\\]/.*/Length [0-9]+( [0-9]+ R)?/.*/Filter/FlateDecode>>")
	byteWidthEx, _ := regexp.Compile("W\\[[0-9] [0-9] [0-9]\\]")
	indexEx, _ := regexp.Compile("Index\\[[0-9]+ [0-9]+\\]")
	numberEx, _ := regexp.Compile("[0-9]")
	reader := bufio.NewReader(toRead)
	id, data, header, err := findBlock(reader, nil, headerEx)
	if err != nil {
		return nil, err
	}
	text, err := uncompress(data)
	if err != nil {
		return nil, err
	}
	rawByteWidth := byteWidthEx.FindString(header)
	rawByteWidths := numberEx.FindAllString(rawByteWidth, 3)
	state, err := strconv.ParseUint(rawByteWidths[0], 10, 64)
	if err != nil {
		return nil, err
	}
	offset, err := strconv.ParseUint(rawByteWidths[1], 10, 64)
	if err != nil {
		return nil, err
	}
	index, err := strconv.ParseUint(rawByteWidths[2], 10, 64)
	if err != nil {
		return nil, err
	}
	xRef := BuildXRef(int(state), int(offset), int(index), text)
	xRef.ID = id
	rawIndexInfo := indexEx.FindString(header)
	rawIndicies := numberEx.FindAllString(rawIndexInfo, 2)
	minIndex, err := strconv.ParseInt(rawIndicies[0], 10, 32)
	if err != nil {
		return nil, err
	}
	maxIndex, err := strconv.ParseInt(rawIndicies[1], 10, 32)
	if err != nil {
		return nil, err
	}
	xRef.MinIndex = int(minIndex)
	xRef.MaxIndex = int(maxIndex)
	return xRef, nil
}

