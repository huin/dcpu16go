package main

import (
	"bufio"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/huin/dcpu16go/core"
)

var (
	flagBigEndian = flag.Bool(
		"big-endian", false,
		"Specifies input is big-endian (little endian is the default).")
)

type ReaderWordLoader struct {
	ByteOrder binary.ByteOrder
	Reader    io.Reader
}

func (r *ReaderWordLoader) WordLoad() (model.Word, error) {
	var word model.Word
	err := binary.Read(r.Reader, r.ByteOrder, &word)
	return word, err
}

func (r *ReaderWordLoader) SkipWords(model.Word) error {
	// Shouldn't be required here.
	panic("unexpected ReaderWordLoader SkipWords call")
}

func main() {
	flag.Parse()

	if flag.NArg() != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <infile> <outfile>\n", os.Args[0])
		flag.PrintDefaults()
	}

	infile, err := os.Open(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	defer infile.Close()

	var byteOrder binary.ByteOrder = binary.LittleEndian
	if *flagBigEndian {
		byteOrder = binary.BigEndian
	}
	wordLoader := &ReaderWordLoader{byteOrder, bufio.NewReader(infile)}

	outfile, err := os.Create(flag.Arg(1))
	if err != nil {
		log.Fatal(err)
	}
	defer outfile.Close()

	var instructionSet model.BasicInstructionSet

	for {
		instruction, err := model.InstructionLoad(wordLoader, &instructionSet)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				log.Fatal(err)
			}
		}
		fmt.Fprintln(outfile, instruction)
	}
}
