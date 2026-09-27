package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

type ValueType string

const (
	ARRAY  ValueType = "*"
	BULK   ValueType = "$"
	STRING ValueType = "+"
	ERROR  ValueType = "-"
)

type Value struct {
	typ   ValueType
	bulk  string
	str   string
	array []Value
}

func readLine(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadString(('\n'))
	if err != nil {
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}

func (v *Value) readArray(reader *bufio.Reader) error {
	line, err := readLine(reader)
	if err != nil {
		return err
	}
	arrlen, err := strconv.Atoi(line[1:])
	if err != nil {
		fmt.Println(err)
		return err
	}

	for range arrlen {
		bulk, err := v.readBulk(reader)
		if err != nil {
			return err
		}
		v.array = append(v.array, bulk)
	}
	return nil
}

func (v *Value) readBulk(reader *bufio.Reader) (Value, error) {
	line, err := readLine(reader)
	if err != nil {
		return Value{}, err
	}

	n, err := strconv.Atoi(line[1:])
	if err != nil {
		fmt.Println(err)
		return Value{}, err
	}

	bulkBuf := make([]byte, n+2)
	_, err = reader.Read(bulkBuf)
	if err != nil {
		return Value{}, err
	}
	return Value{typ: BULK, bulk: string(bulkBuf[:n])}, nil
}

func main() {
	l, err := net.Listen("tcp", ":6379")
	if err != nil {
		log.Fatal("unable to listen")
	}
	defer l.Close()

	conn, err := l.Accept()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)
	for {
		v := Value{typ: ARRAY}
		err := v.readArray(reader)
		if err != nil {
			fmt.Println("read error", err)
			return
		}

		fmt.Println(v.array)
		conn.Write([]byte("+OK\r\n"))
	}

}
