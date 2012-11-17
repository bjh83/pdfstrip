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

func getID(line string) (bool, int64) {
	idEx, _ := regexp.Compile("[0-9]+ [0-9]+ obj")
	rawID := idEx.FindString(line)
	if rawID == "" {
		return false, -1
	}
	return true, getInt64(rawID)
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

