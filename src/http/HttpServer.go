package http

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"strings"
)

type HttpServer struct {
	FilePath string
	Port     string
}

type HttpRequest struct {
	Method string
	Path   string
	Header map[string]string
	Data   string
}

type HttpResponse struct {
	Code   string
	Header map[string]string
	Data   []byte
}

func (server *HttpServer) Serve() {
	ln, err := net.Listen("tcp", ":"+server.Port)
	if err != nil {
		panic(err)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatal("Connect accept error !", err)
		}
		go handleConnect(conn, server)
	}
}

func handleConnect(conn net.Conn, server *HttpServer) {
	request := getHttpRuest(conn)
	response := handlerHttpRequest(request, server)
	writeToResponse(conn, response)
	conn.Close()
}

func getHttpRuest(conn net.Conn) *HttpRequest {
	buffer := bufio.NewReader(conn)
	firstLine, _ := buffer.ReadString('\n')
	firstLine = strings.TrimSpace(firstLine)
	firstLine = strings.Trim(firstLine, "\n")
	if firstLine == "" {
		panic("empty first line")
	}
	method, path := handleHttpFirstLine(firstLine)
	header := make(map[string]string)
	size := buffer.Buffered()
	for size > 0 {
		line, _ := buffer.ReadString('\n')
		line = strings.Trim(line, "\n")
		line = strings.TrimSpace(line)
		if line == "" {
			size = buffer.Buffered()
			break
		}
		paramList := strings.Split(line, ":")
		if len(paramList) < 2 {
			panic("error1!")
		}
		header[paramList[0]] = strings.Join(paramList[1:], ":")
		size = buffer.Buffered()
	}
	data := ""
	if size != 0 {
		buf := make([]byte, size)
		chunks := make([]byte, 1024, 1024)
		for {
			n, err := buffer.Read(buf)
			fmt.Println(buf)
			if err != nil && err != io.EOF {
				panic(err)
			}
			if n == 0 {
				break
			}
			chunks = append(chunks, buf[:n]...)
		}
		data = string(chunks)
	}
	request := &HttpRequest{method, path, header, data}
	return request
}

func handlerHttpRequest(reqeust *HttpRequest, server *HttpServer) *HttpResponse {
	response := new(HttpResponse)
	if reqeust.Path == "" {
		response.Code = "500"
		return response
	}
	header := make(map[string]string)
	header["server"] = "wenlong.bian server"
	filePath := server.FilePath + reqeust.Path
	buf, err := ioutil.ReadFile(filePath)
	if err != nil {
		response.Code = "404"
		response.Header = header
		response.Data = make([]byte, 0)
	}

	nameArray := strings.Split(filePath, ".")
	fileType := nameArray[len(nameArray)-1]
	if fileType == "jpg" {
		header["Content-type"] = "image/jpeg"
	} else if fileType == "pdf" {
		header["Content-type"] = "application/pdf"
	}
	response.Code = "200"
	response.Data = buf
	response.Header = header
	return response
}

func writeToResponse(conn net.Conn, response *HttpResponse) {
	wb := bufio.NewWriter(conn)
	writeData := make([]byte, 0)
	firstLine := "HTTP/1.1 " + response.Code
	code := response.Code
	if code == "200" {
		firstLine += " OK\n"
	} else if code == "404" {
		firstLine += " NOT FOUND\n"
	} else if code == "500" {
		firstLine += " ERROR\n"
		writeData = append(writeData, []byte(firstLine)...)
		wb.Flush()
		return
	}
	writeData = append(writeData, []byte(firstLine)...)
	header := response.Header
	for key, value := range header {
		writeData = append(writeData, []byte(key+": "+value+"\n")...)
	}
	writeData = append(writeData, []byte("\n")...)
	writeData = append(writeData, response.Data...)
	_, err := wb.Write(writeData)
	if err != nil {
		panic(err)
	}
	fmt.Println("wb.Flush()")
}
func handleHttpFirstLine(line string) (method string, path string) {
	firstList := strings.Fields(line)
	if len(firstList) != 3 {
		panic("error2 !")
	}
	method = firstList[0]
	path = firstList[1]
	return
}
