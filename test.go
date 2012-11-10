package main

import(
	"os"
	"io"
	"bufio"
	"fmt"
	"pdfstrip/decode"
	"pdfstrip/deformat"
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
	reader, fileErr := decode.Decode(fileIn)
	if fileErr != nil {
		fmt.Println(fileErr.Error())
		return
	}
	bytebuffer := make([]byte, 1024)
	for {
		num, fileErr := reader.Read(bytebuffer)
		if fileErr != nil && fileErr != io.EOF {
			fmt.Println(fileErr.Error())
			return
		}
		_, fileErr = fileOut.Write(bytebuffer[:num])
		if fileErr != nil {
			fmt.Println(fileErr.Error())
			return
		}
		if num == 0 {
			break
		}
	}
	fileIn.Close()
	fileOut.Close()
	fileIn, fileErr = os.Open(os.Args[2])
	if fileErr != nil {
		fmt.Println(fileErr.Error())
		return
	}
	fileOut, fileErr = os.Create("new_" + os.Args[2])
	if fileErr != nil {
		fmt.Println(fileErr.Error())
		return
	}
	newText, fileErr := deformat.Deformat(fileIn)
	if fileErr != nil {
		fmt.Println(fileErr.Error())
		return
	}
	writer := bufio.NewWriter(fileOut)
	_, fileErr = writer.WriteString(newText)
	fileErr = writer.Flush()
	if fileErr != nil {
		fmt.Println(fileErr.Error())
		return
	}
}

