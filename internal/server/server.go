package server

import (
	"bytes"
	"io"
	"log"
	"net"
	"strconv"
	"sync/atomic"

	"github.com/finfreezer/httpfromtcp/internal/request"
	"github.com/finfreezer/httpfromtcp/internal/response"
)

type Handler func(w *response.Writer, req *request.Request)

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

type Server struct {
	Addr     string
	Handler  Handler
	Listener net.Listener
	IsAlive  *atomic.Bool
	Writer   *response.Writer
}

type ByteReader struct {
	Reader *bytes.Reader
	Closer io.Closer
}

func Serve(port int, handler Handler) (*Server, error) {
	portStr := ":" + strconv.Itoa(port)
	listener, err := net.Listen("tcp", portStr)
	Status := atomic.Bool{}
	Status.Store(true)
	if err != nil {
		log.Printf("Error: %s", err)
	}
	newServer := Server{Addr: portStr, Handler: handler, Listener: listener, IsAlive: &Status, Writer: response.NewResponseWriter()}
	go newServer.listen()
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
	for {
		conn, err := s.Listener.Accept()
		if err != nil {
			if !s.IsAlive.Load() {
				log.Print("Trying to listen while server is closed.")
				return
			}
			log.Printf("Error accepting connection: %v", err)
			continue
		}
		go s.handle(conn)
	}
}

// Use io.NopCloser next time.
func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	s.Writer.ContentWriter = conn
	request, err := request.RequestFromReader(conn)
	if err != nil {
		log.Println("Error requesting from reader")
		return
	}
	log.Println("Finished parsing request.")
	s.Handler(s.Writer, request)
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

/*func (hErr *HandlerError) writeError(w *response.Writer, h headers.Headers) error {
	err := w.WriteStatusLine(hErr.StatusCode)
	if err != nil {
		log.Println("Error in writeError.")
		return err
	}
	err = w.WriteHeaders(h)
	if err != nil {
		log.Println("Error in writeError.")
		return err
	}
	_, err = w.WriteBody([]byte(hErr.Message))
	if err != nil {
		log.Println("Error in writeError.")
		return err
	}
	log.Println(w.ResponseHTML)
	return nil
}*/
