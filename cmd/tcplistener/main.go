package main

import (
	"fmt"
	"log"
	"net"

	"httpfromtcp/internal/request"
)

func main() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatal("Err: ", err)
	}
	defer listener.Close()
	fmt.Println("listening on port :42069")
	// os.Stdout.Sync()   // flush stdout

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("Error accepting: ", err)
		}
		fmt.Println("Connection accepted.")
		r, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Request line:")
		fmt.Println("- Method:", r.RequestLine.Method)
		fmt.Println("- Target:", r.RequestLine.RequestTarget)
		fmt.Println("- Version:", r.RequestLine.HttpVersion)

		fmt.Println("Headers:")
		r.Headers.ForEach(func(n, v string) {
			fmt.Printf("- %s: %s\n", n, v)
		})

		fmt.Println("Body:")
		fmt.Printf("%s", r.Body)
		conn.Close()
	}
}
