package main

import(
	"os"
	"fmt"
	"pdfstrip/decode"
	"pdfstrip/deformat"
	"encoding/xml"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Please provide two file names...")
		return
	}
	fileIn, fileErr := os.Open(os.Args[1])
	if fileErr != nil {
		fmt.Println(fileErr.Error())
		return
	}
	fileOut, fileErr := os.Create(os.Args[2])
	if fileErr != nil {
		fmt.Println(fileErr.Error())
		return
	}
	fileData, fileErr := decode.Decode(fileIn)
	if fileErr != nil {
		fmt.Println(fileErr.Error())
		return
	}
	byteBuffer, err := xml.Marshal(fileData)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	_, fileErr = fileOut.Write(byteBuffer)
	fileIn.Close()
	fileOut.Close()
	fileOut, fileErr = os.Create("new_" + os.Args[2])
	if fileErr != nil {
		fmt.Println(fileErr.Error())
		return
	}
	for index, val := range fileData.Blocks {
		fileData.Blocks[index].Text, _ = deformat.Deformat([]byte(val.Text))
	}
	byteBuffer, err = xml.Marshal(fileData)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	_, fileErr = fileOut.Write(byteBuffer)
	if fileErr != nil {
		fmt.Println(fileErr.Error())
		return
	}
}

