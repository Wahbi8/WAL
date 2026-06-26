package main

type block struct {
	header header
	record string // i might change this to byte in the future
}

type header struct{
	checkSum uint32 
	logType uint8
	lenght uint32
}

func main()  {

}