package deformat

import(
	"io"
	"bufio"
	"regexp"
	"strconv"
	"strings"
)

const(
	textLine = iota
	coordLine = iota
	other = iota
)

type Document struct {
	Name string
	Body []Page
}

type Page struct {
	Number int
	Text string
}

func Deformat(toRead io.Reader) (string, error) {
	var page string
	var currentPos float32
	var lastPos float32
	reader := bufio.NewReader(toRead)
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		switch getLineType(line) {
		case textLine:
			if currentPos < lastPos {
				page += "\n"
			}
			lastPos = currentPos
			page += stripText(line)
		case coordLine:
			currentPos = getXPos(line)
		case other:
		}
	}
	return page, nil
}

func getLineType(line string) int {
	text, _ := regexp.Compile("\\(\\.+\\)T[jJ]")
	coord, _ := regexp.Compile("1 0 0 1 [0-9]+(\\.[0-9]+)? [0-9]+(\\.[0-9]+)?")
	if text.MatchString(line) {
		return textLine
	} else if coord.MatchString(line) {
		return coordLine
	} else {
		return other
	}
	return other
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
	regex, _ := regexp.Compile("\\(\\.+\\)T[jJ]")
	text := regex.FindString(line)
	text = text[1:len(text)-3]
	return text
}

