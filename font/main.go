package main

import (
	"fmt"
	"os"
)

const (
	normalFont = 0xdb20 - 0xc000
	boldFont   = 0xdd60 - 0xc000
)

func main() {
	rom := os.Args[1]

	data, _ := os.ReadFile(rom)

	dumpFont(data, normalFont, 6)
	dumpFont(data, boldFont, 7)
}

func dumpFont(data []byte, offset int, l int) {
	fmt.Printf("Dumping font at 0x%04X len: %d\n", offset, l)
	for i := 0; i < 96; i++ {
		chr := byte(0x20 + i)
		fmt.Printf("\nGlyph: %c\n", chr)
		g := make([]byte, l)
		for j := 0; j < l; j++ {
			pos := offset + (i * l) + j
			g[j] = data[pos]
		}
		for j := 0; j < 8; j++ {
			var mask byte = 1 << j
			for _, b := range g {
				if b&mask != 0 {
					fmt.Printf("*")
				} else {
					fmt.Printf(" ")
				}
			}
			fmt.Printf("\n")
		}
	}
}
