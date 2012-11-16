package main

import(
	"os"
	"encoding/xml"
	"fmt"
	"pdfstrip/decode"
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
	xRef, fileErr := decode.GetXRef(fileIn)
	if fileErr != nil {
		fmt.Println(fileErr.Error())
		return
	}
	data, fileErr := xml.MarshalIndent(xRef, "\t", "\t")
	if fileErr != nil {
		fmt.Println(fileErr.Error())
		return
	}
	_, fileErr = fileOut.Write(data)
	if fileErr != nil {
		fmt.Println(fileErr.Error())
		return
	}
}

