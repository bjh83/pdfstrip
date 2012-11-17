package main

import(
	"os"
	"fmt"
	"pdfstrip/decode"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Please provide one file names...")
		return
	}
	buffer := make([]byte, 32)
	fileIn, fileErr := os.Open(os.Args[1])
	if fileErr != nil {
		fmt.Println(fileErr.Error())
		return
	}
	xRef, fileErr := decode.GetXRef(fileIn)
	if fileErr != nil {
		fmt.Println(fileErr.Error())
		return
	}
	offset := int64(xRef.Trips[43].Offset)
	_, fileErr = fileIn.Seek(offset, 0)
	if fileErr != nil {
		fmt.Println(fileErr.Error())
		return
	}
	_, fileErr = fileIn.Read(buffer)
	if fileErr != nil {
		fmt.Println(fileErr.Error())
		return
	}
	fmt.Println(string(buffer))
}

