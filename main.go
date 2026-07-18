package main

import (
	"encoding/binary"
	"hash/crc32"
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
	recordId 	uint16
	checkSum 	uint32	// 4 bytes	- fingerprint of the payload
	logType 	uint8	// 1 byte	- type is (full / start / middle / last) aa a number
	lenght 		uint16	// 2 bytes	- how many bytes is the payload
	payload 	[]byte  // operation -> keyLength -> key -> value

	payloadStruct payload
}
// i need to add an identifier for type
type payload struct{
	operation 	uint8
	keyLength	uint16
	valueLength	uint16
	key			[]byte
	value		[]byte
}

type FragmentReassembler struct {
    buffers map[uint16]*tempRecord 
}

type tempRecord struct {
	// expectedLen uint32
	// recievedLen uint32
	data []byte
}

func main() {

	
}

const blockSize = 32 * 1024
const headerSize = 7
const maxPayloadSize = blockSize - headerSize

func (r *record) serialize() []byte {

	// i need to chnage this so it will not exceeds 32KB -> 32768 bytes

	if len(r.payload) <= maxPayloadSize {
		r.checkSum = crc32.ChecksumIEEE(r.payload)
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

func deserializeHeader(bytes []byte) record {

	return record{
		checkSum: binary.LittleEndian.Uint32(bytes[0:4]),
		logType: bytes[4],
		lenght: binary.LittleEndian.Uint16(bytes[5:7]),
		payloadStruct: payload{
			operation: bytes[7],
			keyLength: binary.LittleEndian.Uint16(bytes[8:10]),
			valueLength: binary.LittleEndian.Uint16(bytes[10:12]),
		},
	}
}

func compareCheckSum(headerCheckSum uint32, payload []byte) bool {

	payloadChechSum := crc32.ChecksumIEEE(payload)
	return headerCheckSum == payloadChechSum
}

// implement fragment reassembly (first, middle, last)
// i'll need a struct that will hold the record content temporarily 
// i need gloo the the parts and return record
// i need to use the lengths to know where the key ends
// fr is a buffer
func (fr *FragmentReassembler) Assemble(r record) (record, bool) {

	switch r.logType {
	case uint8(full):
		return r, true
	case uint8(start):
		fr.buffers[r.recordId] = &tempRecord{
			data: append([]byte(nil), r.payload...),
		}
		return r, false
	case uint8(middle):
		if d, ok := fr.buffers[r.recordId]; ok {
			d.data = append(d.data, r.payload...)
		}  
		return  r, false
	case uint8(end):
		if d, ok := fr.buffers[r.recordId]; ok {
			d.data = append(d.data, r.payload...)
			
			return parseRecord(d.data), true
		}
		return record{}, false
	}

	return record{}, false
}

func parseRecord(data []byte) record {


}