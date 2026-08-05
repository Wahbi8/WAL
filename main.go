package main

import (
	"encoding/binary"
	"hash/crc32"
	"os"
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
	recordId 	uint16	// 2 bytes
	checkSum 	uint32	// 4 bytes	- fingerprint of the payload
	logType 	uint8	// 1 byte	- type is (full / start / middle / last) aa a number
	length		uint32	// 4 bytes	- how many bytes is the payload
	payload 	[]byte  // operation -> keyLength -> key -> value

	payloadStruct payload
}
// i need to add an identifier for type
type payload struct{
	operation 	uint8
	keyLength	uint16
	valueLength	uint32
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

type Store struct {
	file *os.File
	lastRecordId uint16
}

func main() {

	
}

const blockSize = 32 * 1024
const headerSize = 11
const maxPayloadSize = blockSize - headerSize

func OpenFile(path string) (*Store, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	// todo: read last record id
	return &Store{file: f}, nil
}

func (s *Store) writeRecord(data []byte) error {

	if _, err := s.file.Write(data); err != nil {
		return err
	}
	return nil
}

func (s *Store) nextId() uint16 {
	s.lastRecordId++
	return s.lastRecordId
}
 
func (s *Store) serialize(payload []byte) error {

	// i need to chnage this so it will not exceeds 32KB -> 32768 bytes

	// add recordId by reading the last record id + 1 or 1 if there is none

	if len(payload) <= maxPayloadSize {
		r := record{
			recordId: s.nextId(),
			checkSum: crc32.ChecksumIEEE(payload),
			logType:  uint8(full),
			length:   uint32(len(payload)),
			payload:  payload,
		}
		return s.serializeRecord(r) // i need to store in disk
	}

	num := 0
	globalRecordId := s.nextId()

	for num < len(payload) {

		endPayload := num + maxPayloadSize
		if endPayload > len(payload) {
			endPayload = len(payload)
		}

		var typeRecord byte
		switch {
		case num == 0:
			typeRecord = byte(start) // start
		case endPayload == len(payload):
			typeRecord = byte(end) // end
		default:
			typeRecord = byte(middle) // middle
		}

		chunk := payload[num:endPayload]

		r := record{
			recordId: globalRecordId,
			checkSum: crc32.ChecksumIEEE(chunk),
			logType:  typeRecord,
			length:   uint32(len(chunk)),
			payload:  chunk,
		}

		if err := s.serializeRecord(r); err != nil {
			return err
		}
		num = endPayload
	}
	return nil
}

func (s *Store) serializeRecord(r record) error {

	totalSize := headerSize + len(r.payload)
	buf := make([]byte, totalSize)

	binary.LittleEndian.PutUint16(buf[0:2], r.recordId)
	binary.LittleEndian.PutUint32(buf[2:6], r.checkSum)
	buf[6] = r.logType
	binary.LittleEndian.PutUint32(buf[7:11], r.length)
	copy(buf[11:], r.payload)

	if _, err := s.file.Write(buf); err != nil {
		return err
	}

	return nil
}

func deserializeHeader(bytes []byte) record {

	return record{
		recordId: binary.LittleEndian.Uint16(bytes[0:2]),
		checkSum: binary.LittleEndian.Uint32(bytes[2:6]),
		logType: bytes[6],
		length: binary.LittleEndian.Uint32(bytes[7:11]),
		payloadStruct: payload{
			operation: bytes[11],
			keyLength: binary.LittleEndian.Uint16(bytes[12:14]),
			valueLength: binary.LittleEndian.Uint32(bytes[14:18]),
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
			// i need to empty the buffer 'fr'
			return parseRecord(d.data, r), true
		}
		return record{}, false
	}

	return record{}, false
}

func parseRecord(data []byte, r record) record {

	operation := logType(data[0])
	data = data[1:]

	kLength := binary.LittleEndian.Uint16(data[0:2])
	data = data[2:]

	vLength := binary.LittleEndian.Uint32(data[0:4])
	data = data[4:]

	kValue := data[0:kLength]
	data = data[kLength:]

	vValue := data[0:vLength]
	data = data[vLength:]

	// planning to have one key per record. (code stays untill i change my mind)
	for len(data) > 0 {
		data = data[1:]			// pop operation byte[0]
		
		tempKLength := binary.LittleEndian.Uint16(data[0:2])
		data = data[2:]			// pop key byte[0:2]

		tmpVLength := binary.LittleEndian.Uint32(data[0:4])
		data = data[4:]

		data = data[tempKLength:]	// pop key value

		vValue = append(vValue, data[0:tmpVLength]...)
		data = data[tmpVLength:]
	}

	vl := len(vValue)
	var pl []byte

	pl = append(pl, byte(operation))
	pl = binary.LittleEndian.AppendUint16(pl, kLength)		// to insure 2 bytes even if it is smaller
	pl = binary.LittleEndian.AppendUint32(pl, uint32(vl))	// to insure 4 bytes
	pl = append(pl, kValue...)
	pl = append(pl, vValue...)

	return record{
		recordId: r.recordId,
		checkSum: r.checkSum,
		logType: uint8(full),
		length: uint32(7 + len(kValue) + len(vValue)), // 7 is the header bytes (op + klen + vlen)
		payload: pl,

	}
}