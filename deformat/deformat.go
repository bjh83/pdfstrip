package deformat

import(
	"io"
	"bufio"
	"bytes"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf16"
	"pdfstrip/decode"
)

type Document struct {
	Name string
	Body []Page
}

type Page struct {
	Number int
	Text string
}

func Deformat(data []byte) (string, error) {
	var page string
	var currentPos float32
	var lastPos float32
	reader := bufio.NewReader(bytes.NewBuffer(data))
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		textLine, coordLine := getLineType(line)
		if coordLine {
			currentPos = getXPos(line)
		}
		if textLine {
			if currentPos < lastPos {
				page += "\n"
			}
			lastPos = currentPos
			text := stripText(line)
			page += text
		}
	}
	return page, nil
}

func getLineType(line string) (bool, bool) {
	text, _ := regexp.Compile("\\(.+\\)T[jJ]")
	coord, _ := regexp.Compile("1 0 0 1 [0-9]+(\\.[0-9]+)? [0-9]+(\\.[0-9]+)?")
	textLine, coordLine := false, false
	if text.MatchString(line) {
		textLine = true
	}
	if coord.MatchString(line) {
		coordLine = true
	}
	return textLine, coordLine
}

func getXPos(line string) float32 {
	coordString := getCoordinateString(line)
	coordinates := strings.Split(coordString, " ")
	xCoord := coordinates[4]
	xFloat, _ := strconv.ParseFloat(xCoord, 32)
	return float32(xFloat) //We do not really care if it fails,
	// it will just start a new line
}

func getCoordinateString(line string) string {
	regex, _ := regexp.Compile("1 0 0 1 [0-9]+(\\.[0-9]+)? [0-9]+(\\.[0-9]+)?")
	return regex.FindString(line)
}

func stripText(line string) string {
	regex, _ := regexp.Compile("\\(.+\\)T[jJ]")
	text := regex.FindString(line)
	text = text[1:len(text)-3]
	return text
}

func decode(stream []byte) []rune {
	out := make([]rune, len(stream) / 2)
	for i := 1; i < len(stream); i += 2 {
		out[i / 2] = utf16.DecodeRune(rune(stream[i - 1]), rune(stream[i]))
	}
	return out
}

