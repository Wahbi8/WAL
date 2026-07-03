package main

import (
	"encoding/binary"
)

type logType int

const (
	full logType = iota
	start
	middle
	end
)

// block size 32KB
type record struct{
	checkSum uint32	// 4 bytes	- fingerprint of the payload
	logType uint8	// 1 byte	- type is (full / start / middle / last) aa a number
	lenght uint16	// 2 bytes	- how many bytes is the payload
	payload []byte  // operation -> keyLength -> key -> value
}
// i need to add an identifier for type

func main() {

	
}

const blockSize = 32 * 1024
const headerSize = 7
const maxPayloadSize = blockSize - headerSize

func (r *record) serialize() []byte {

	// i need to chnage this so it will not exceeds 32KB -> 32768 bytes

	if len(r.payload) <= maxPayloadSize {
		// r.checkSum = ,
		r.lenght = uint16(len(r.payload))
		r.logType = uint8(full) // full
		return serializeRecord(*r)
	}

	var out []byte

	num := 0
	for num < len(r.payload) {

		end := num + maxPayloadSize
		if end > len(r.payload) {
			end = len(r.payload)
		}

		var typeRecord byte
		switch {
		case num == 0:
			typeRecord = byte(start) // start
		case end == len(r.payload):
			typeRecord = byte(end) // end
		default:
			typeRecord = byte(middle) // middle
		}

		payloadPart := r.payload[num:end]

		recordPart := record{
			// checkSum: ,
			logType: typeRecord,
			lenght: uint16(len(payloadPart)),
			payload: payloadPart,
		}
		out = append(out, serializeRecord(recordPart)...)
		num = end
	}
	return out
}

func serializeRecord(r record) []byte {

	totalSize := 7 + len(r.payload)
	buf := make([]byte, totalSize)

	binary.LittleEndian.PutUint32(buf[0:4], r.checkSum)
	buf[4] = r.logType
	binary.LittleEndian.PutUint16(buf[5:7], r.lenght)
	copy(buf[7:], r.payload)

	return buf
}