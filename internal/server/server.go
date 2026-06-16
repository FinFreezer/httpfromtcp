package server

import (
	"bytes"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"sync/atomic"

	"github.com/finfreezer/httpfromtcp/internal/response"
)

type Server struct {
	Addr     string
	Handler  http.Handler
	Listener net.Listener
	IsAlive  *atomic.Bool
}

type ByteReader struct {
	Reader *bytes.Reader
	Closer io.Closer
}

func Serve(port int) (*Server, error) {
	portStr := ":" + strconv.Itoa(port)
	listener, err := net.Listen("tcp", portStr)
	Status := atomic.Bool{}
	Status.Store(true)
	if err != nil {
		log.Printf("Error: %s", err)
	}
	newServer := Server{Addr: portStr, Handler: nil, Listener: listener, IsAlive: &Status}
	go func(newServer *Server) {
		newServer.listen()
	}(&newServer)
	return &newServer, nil
}

func (s *Server) Close() error {
	err := s.Listener.Close()
	if err != nil {
		log.Println("Error closing listener.")
		return err
	}
	s.IsAlive.Swap(false)
	return nil
}

func (s *Server) listen() {
	if !s.IsAlive.Load() {
		log.Print("Trying to listen while server is closed.")
		return
	}
	for {
		conn, err := s.Listener.Accept()
		if err != nil {
			log.Println(err)
		}
		go func(c net.Conn) {
			s.handle(c)
		}(conn)
	}
}

// Use io.NopCloser next time.
func (s *Server) handle(conn net.Conn) {
	err := response.WriteStatusLine(conn, 200)
	if err != nil {
		log.Println(err)
		return
	}
	h := response.GetDefaultHeaders(0)
	err = response.WriteHeaders(conn, h)
	if err != nil {
		log.Println(err)
		return
	}
	/*h := http.Header{}
	h.Add("Content-Type", "text/plain")
	respBody := []byte("\nHello World!")
	newReader := bytes.NewReader(respBody)
	bodyReader := ByteReader{Reader: newReader}
	newRespPlain := []byte("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\n\r\nHello World!")
	newResp := http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Proto: "HTTP/1.1",
		Header:        h,
		Body:          bodyReader,
		ContentLength: 13,
	}*/
	conn.Close()
}

func (b ByteReader) Close() error {
	if b.Closer != nil {
		return b.Closer.Close()
	}
	return nil
}

func (b ByteReader) Read(data []byte) (int, error) {
	return b.Reader.Read(data)
}
