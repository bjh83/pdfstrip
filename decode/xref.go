package decode

import(
	"io"
	"io/ioutil"
	"compress/flate"
	"regexp"
	"strconv"
	"errors"
)

var UnrecognizedFormat error = errors.New("Does not match known format")
var XRefError error = errors.New("Did not find expected object")

func Uncompress(stream []byte) ([]byte, error) {
	stream = append(Dictionary, stream...)
	uncompressed, err := ioutil.ReadAll(flate.NewReader(bytes.NewBuffer(stream)))
	if err != nil {
		return nil, err
	}
	return uncompressed, nil
}

func getInt64(line string) int64 {
	numberEx, _ := regexp.Compile("[0-9]+")
	rawNumber := numberEx.FindString(line)
	out, _ := strconv.ParseInt(rawNumber, 10, 64)
	return out
}

func getLength(line string) (bool, int64) {
	lengthEx, _ := regexp.Compile("/Length [0-9]+")
	length := lengthEx.FindString(line)
	if length == "" {
		return false, -1
	}
	return true, getInt64(length)
}

func getLengthByID(id int, file io.ReadSeeker, xRef *XRefBlock) (int64, error) {
	length, hasLength := lengthHash[id]
	if hasLength {
		return length, nil
	}
	saveOffset, err := file.Seek(0, 1)
	if err != nil {
		return -1, err
	}
	newOffset := xRef.GetOffset(id)
	_, err := file.Seek(newOffset, 0)
	_, err := util.ReadLine(file)
	rawLength := util.Readline(file)
	_, err := file.Seek(saveOffset, 0)
	return getInt64(rawLength), nil
}

func getLength(line string, file io.ReadSeeker, xRef *XRefBlock) (bool, int64) {
	lengthEx, _ := regexp.Compile("/Length [0-9]+ [0-9]+ R")
	rawLength := lengthEx.FindString(line)
	if rawLength == "" {
		return getLength(line)
	}
	id := getInt64(rawLength)
	length, err := getLengthByID(id, file, xRef)
	if err != nil {
		return false, -1
	}
	return true, length
}

func getID(line string) (bool, int64) {
	idEx, _ := regexp.Compile("[0-9]+ [0-9]+ obj")
	rawID := idEx.FindString(line)
	if rawID == "" {
		return false, -1
	}
	return true, getInt64(rawID)
}

func findBlock(file io.ReadSeeker, id int64, xRef *XRefBlock) (string, []byte, error) {
	offset := xRef.GetOffset(id)
	_, err := file.Seek(offset, 0)
	if err != nil {
		return "", nil, err
	}
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", nil, err
	}
	hasID, toCompare := getID(line)
	if !hasID || toCompare != id {
		return nil, XRefError
	}
	header := util.GetHeader(file)
	length := getLength(header, xRef)
	buffer := make([]byte, length)
	_, err := file.Read(buffer)
	if err != nil {
		return "", nil, err
	}
	return header, buffer, nil
}

func findXRef(reader bufio.Reader) (int, []byte, string, error) {
	headerEx, _ := regexp.Compile("<</Type/XRef/.*/Filter/FlateDecode>>")
	startEx, _ := regexp.Compile("<<")
	endEx, _ := regexp.Compile(">>stream")
	var id int64
	shouldAppend := false
	header := ""
	for {
		line, err := reader.ReadString('\n')
		line = strings.Replace(line, "\n", "" -1)
		if err != nil {
			return 0, nil, "", err
		}
		hasID, tempID := getID(line)
		if hasID {
			id = tempID
		}
		if startEx.MatchString(line) {
			header = line
			shouldAppend = true
		}
		if shouldAppend {
			header += line
			if endEx.MatchString(header) {
				if headerEx.MatchString(header) {
					break
				} else {
					shouldAppend = false
					header = ""
				}
			}
		}
	}
	hasLength, length := getLength(header)
	if !hasLength {
		return -1, nil, "", UnrecognizedFormat
	}
	data := make([]byte, length)
	_, err := reader.Read(data)
	if err != nil {
		return -1, nil, "", err
	}
	return id, data, header, nil
}

func GetXRef(toRead io.Reader) (*XRefBlock, error) {
	byteWidthEx, _ := regexp.Compile("W\\[[0-9] [0-9] [0-9]\\]")
	indexEx, _ := regexp.Compile("Index\\[[0-9]+ [0-9]+\\]")
	numberEx, _ := regexp.Compile("[0-9]+")
	reader := bufio.NewReader(toRead)
	id, data, header, err := findXRef(reader)
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

