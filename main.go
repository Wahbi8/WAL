package main

import (
	"encoding/binary"
)

type block struct {
	header header
	record record
}

type header struct{
	checkSum uint32	// 4 bytes
	logType uint8	// 1 byte	// type is (full / start / middle / last) aa a number
	lenght uint32	// 2 bytes
}

type record struct{
	operation uint8		// 1 byte  // enum (delete / update)
	keyLength uint8		// 2 bytes 
	valueLength uint32	// 4 bytes
	key []byte			// 2 bytes
	value []byte		// 16 bytes
}

func main() {

	
}

func (b *block) serialize() []byte {

}