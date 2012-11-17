package decode

import(
	"io"
	"regexp"
	"strings"
	"bufio"
)

func GetRoot(file io.ReadSeeker, offset, id int64) (RootBlock, error) {
	_, err := file.Seek(offset, 0)
	reader := bufio.NewReader(file)
	line, err := reader.ReadString('\n')
	hasID, toCompare := getID(line)
	if !hasID || toCompare != id {
		return RootBlock{}, XRefError
	}
	endEx, _ := regexp.Compile("endobj")
	var buffer string
	for {
		line, err := reader.ReadString('\n')
		line = strings.Replace(line, "\n", "", -1)
		if err != nil {
			return RootBlock{}, err
		}
		if endEx.MatchString(line) {
			return parseRoot(buffer), nil
		}
		buffer += line
	}
}

func parseRoot(buffer string) RootBlock {
	block := RootBlock{}
	treeRootEx, _ := regexp.Compile("StructTreeRoot [0-9]+ [0-9]+ R")
	outlineEx, _ := regexp.Compile("Outlines [0-9]+ [0-9]+ R")
	langEx, _ := regexp.Compile("Lang([a-z]+)")
	metadataEx, _ := regexp.Compile("Metadata [0-9]+ [0-9]+ R")
	pagesEx, _ := regexp.Compile("Pages [0-9]+ [0-9]+ R")
	rawTreeRoot := treeRootEx.FindString(buffer)
	rawOutline := outlineEx.FindString(buffer)
	rawLang := langEx.FindString(buffer)
	rawMetadata := metadataEx.FindString(buffer)
	rawPages := pagesEx.FindString(buffer)
	block.TreeRoot = getInt64(rawTreeRoot)
	block.Outline = getInt64(rawOutline)
	rawLang = strings.Split(rawLang, "(")[0]
	rawLang = strigns.Replace(rawLang, ")", "", -1)
	block.Lang = rawLang
	block.Metadata = getInt64(rawMetadata)
	block.Pages = getInt64(rawPages)
	return block
}

