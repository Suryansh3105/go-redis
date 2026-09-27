package main

import (
	"bufio"
	"fmt"
	"io"
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
	NULL   ValueType = ""
)

type Value struct {
	typ   ValueType
	bulk  string
	str   string
	err   string
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
	writer := NewWriter(conn)
	for {
		v := Value{typ: ARRAY}
		err := v.readArray(reader)
		if err != nil {
			fmt.Println("read error", err)
			return
		}
		handle(writer, &v)

		fmt.Println(v.array)
	}

}

type Handler func(*Value) *Value

var Handlers = map[string]Handler{}

func handle(w *Writer, v *Value) {
	if len(v.array) == 0 {
		w.write(&Value{typ: ERROR, err: "empty command"})
		return
	}
	cmd := v.array[0].bulk
	handler, ok := Handlers[cmd]
	if !ok {
		w.write(&Value{typ: ERROR, err: "unknown command '" + cmd + "'"})
		return
	}
	reply := handler(v)

	w.write(reply)
}

type Writer struct {
	writer *bufio.Writer
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{writer: bufio.NewWriter(w)}
}

func (w *Writer) write(v *Value) error {
	var reply string
	switch v.typ {
	case STRING:
		reply = fmt.Sprintf("%s%s\r\n", v.typ, v.str)
	case BULK:
		reply = fmt.Sprintf("%s%d\r\n%s\r\n", v.typ, len(v.bulk), v.bulk)
	case ERROR:
		reply = fmt.Sprintf("%s%s\r\n", v.typ, v.err)
	case NULL:
		reply = "$-1\r\n"
	}

	_, err := w.writer.Write([]byte(reply))
	if err != nil {
		return err
	}
	return w.writer.Flush()
}
