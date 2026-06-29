package main

import (
	"encoding/binary"
)

// block size 32KB
type record struct{
	checkSum uint32	// 4 bytes	- fingerprint of the payload
	logType uint8	// 1 byte	- type is (full / start / middle / last) aa a number
	lenght uint16	// 2 bytes	- how many bytes is the payload
	payload []byte  // operation -> keyLength -> key -> value
}

func main() {

	
}

func (r *record) serialize() []byte {

	// i need to chnage this so it will not exceeds 32KB -> 32000 bytes
	if len(r.payload) >= 31993 {

		numRecords := len(r.payload) / 31993

		for num := 0; num < numRecords; num++ {
			
		}
	} else {

		return r.serializeRecord()
	}
	
}

func (r *record) serializeRecord() []byte {

	totalSize := 7 + len(r.payload)
	buf := make([]byte, totalSize)

	binary.LittleEndian.PutUint32(buf[0:4], r.checkSum)
	buf[4] = r.logType
	binary.LittleEndian.PutUint16(buf[5:7], r.lenght)
	copy(buf[7:], r.payload)

	return buf
}