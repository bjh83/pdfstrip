package util

import(
	"io"
)

const(bufferSize = 256)

var lastOffset int64
var nextIndex int = bufferSize
var readBuffer []byte = make([]byte, bufferSize)

func ReadLine(file io.ReadSeeker) (string, error) {
	localBuffer := make([]byte, 0)
	offset, err := file.Seek(0, 1)
	if err != nil {
		return "", err
	}
	if lastOffset != offset {
		nextIndex = bufferSize
	}
	for {
		newChar, err := getNext(file)
		if err != nil {
			return "", err
		}
		localBuffer = append(localBuffer, newChar)
		if newChar == '\n' {
			offset += len(localBuffer)
			lastOffset = offset
			_, err = file.Seek(offset, 0)
			if err != nil {
				return "", err
			}
			return string(localBuffer), nil
		}
	}
}

func getNext(file io.ReadSeeker) (byte, error) {
	if nextIndex < bufferSize {
		retVal := readBuffer[nextIndex]
		nextIndex++
		return retVal, nil
	}
	_, err := file.Read(readBuffer)
	if err != nil {
		return byte(0), err
	}
	nextIndex = 1
	return readBuffer[0], nil
}

func GetHeader(file io.ReadSeeker) (string, error) {
}

