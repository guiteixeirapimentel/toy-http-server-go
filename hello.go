package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

func get_404_response() string {

	body_content := "Could not find file!"

	response := "HTTP/1.1 404 Not Found\r\n"
	response += "Connection: Closed\r\n"
	response += "Content-Type: text/html; charset=utf-8\r\n"
	response += "Content-Length: " + strconv.Itoa(len(body_content)) + "\r\n"
	response += "Content-Type: text/html\r\n\r\n"
	response += body_content
	return response
}

func get_file_content(filename string) (string, error) {
	file_handle, err := os.Open(filename)

	if err != nil {
		fmt.Println(err)
		return "", err
	}

	result := ""

	buffer := make([]byte, 4096)

	for {
		nbytes, err := file_handle.Read(buffer)

		if err != nil && nbytes != 0 {
			return "", err
		}
		if nbytes == 0 {
			break
		}
		result += string(buffer[0:nbytes])
	}

	return result, nil
}

func get_filename_from_http_request(request []byte, nbytes int) (string, error) {
	req_as_string := string(request[0:nbytes])

	fmt.Println(req_as_string)

	lines := strings.Split(req_as_string, "\n")
	if len(lines) == 0 {
		return "", errors.New("no lines to parse")
	}
	tokens := strings.Split(lines[0], " ")

	if tokens[0] != "GET" {
		err := "Received invalid method: " + tokens[0]
		return "", errors.New(err)
	}

	if tokens[2] != "HTTP/1.1\r" {
		err := "Received invalid http version: " + tokens[2]
		return "", errors.New(err)
	}

	fmt.Println("Got ", tokens[1])

	return tokens[1], nil

}

func handle_connection(conn net.Conn) {
	defer conn.Close()

	bytes := make([]byte, 1024)
	nbytes, err := conn.Read(bytes)

	if err != nil {
		fmt.Println("Got error ", err)
		return
	}

	if nbytes > 0 {
		fmt.Printf("Got bytes: %s\n", bytes)
	}

	filename, err := get_filename_from_http_request(bytes, nbytes)

	if err != nil {
		fmt.Println("get_filename_from_http_request failed: ", err)
		return
	}

	body_content, err := get_file_content(filename[1:])

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			conn.Write([]byte(get_404_response()))
			return
		}
		fmt.Println("got error ", err)
		return
	}

	response := "HTTP/1.1 200 OK\r\n"
	response += "Connection: Closed\r\n"
	response += "Content-Type: text/html; charset=utf-8\r\n"
	response += "Content-Length: " + strconv.Itoa(len(body_content)) + "\r\n"
	response += "Content-Type: text/html\r\n\r\n"
	response += body_content

	conn.Write([]byte(response))
}

func main() {
	ln, err := net.Listen("tcp", "127.0.0.1:8090")

	if err != nil {
		fmt.Println("Got error: ", err)
		return
	}

	for {
		conn, err := ln.Accept()

		if err != nil {
			fmt.Println("Failed to accept connection: ", err)
			return
		}
		go handle_connection(conn)

		fmt.Println("Handled conn: ", conn)
	}

}
