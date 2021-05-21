package main

import (
	"debug/elf"
	"fmt"
	"log"
	"os"
)

func readProg(ef *elf.File, addr, size uint64) ([]byte, error) {
	if addr == 0 || size == 0 {
		return nil, fmt.Errorf("addr (%v) and size (%v) must both be non-zero", addr, size)
	}

	data := make([]byte, size)
	for _, prog := range ef.Progs {
		if prog.Vaddr <= addr && addr+size-1 <= prog.Vaddr+prog.Filesz-1 {
			_, err := prog.ReadAt(data, int64(addr-prog.Vaddr))
			if err != nil {
				return nil, err
			}
			return data, nil
		}
	}

	return nil, fmt.Errorf("unable to read prog table at addr %v", addr)
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("usage: %v <ELF file>", os.Args[0])
	}

	elfFile, err := elf.Open(os.Args[1])
	if err != nil {
		log.Fatalf("error opening ELF file %v: %v", os.Args[1], err)
	}

	symbols, err := elfFile.Symbols()
	if err != nil {
		log.Fatalf("error reading symbols table from file %v: %v", os.Args[1], err)
	}

	symbol := elf.Symbol{}
	for i := range symbols {
		if symbols[i].Name == "runtime.buildVersion" {
			symbol = symbols[i]
			break
		}
	}

	addr := symbol.Value
	size := symbol.Size
	data, err := readProg(elfFile, addr, size)
	if err != nil {
		log.Fatalf("error in reading address of string: %v", err)
	}

	addr = elfFile.ByteOrder.Uint64(data)
	size = elfFile.ByteOrder.Uint64(data[8:])
	data, err = readProg(elfFile, addr, size)
	if err != nil {
		log.Fatalf("error in reading string: %v", err)
	}

	fmt.Println(string(data))
}
