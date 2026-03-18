package cli

import (
	"fmt"
	"io"
	"iter"
	"os"
	"unicode/utf8"
)

func getRunesFromFile(fileName string) (iter.Seq[rune], int64, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to open file: %w", err)
	}
	fileSize, err := f.Seek(0, io.SeekEnd)
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to seek the end in file: %w", err)
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return nil, 0, fmt.Errorf("Failed to return to start of file: %w", err)
	}
	return func(yield func(rune) bool) {
		defer f.Close()
		const bufferSize = 256
		buffer := new([bufferSize]uint8)
		leftBytes := 0
		totalReadBytes := 0
		for totalReadBytes < int(fileSize) {
			bytesRead, err := f.Read(buffer[leftBytes:])
			if err != nil {
				return
			}
			totalReadBytes += bytesRead
			bytesRead += leftBytes
			i := 0
			for i < bytesRead {
				r, size := utf8.DecodeRune(buffer[i:])
				if r == utf8.RuneError {
					// We assume that we have only part of rune in buffer
					if bytesRead-i > 4 {
						return // Max possible rune size is 4 bytes
					}
					leftBytes = copy(buffer[:], buffer[i:bytesRead])
					break
				}
				if !yield(r) {
					return
				}
				i += size
			}
		}
	}, fileSize, nil
}
