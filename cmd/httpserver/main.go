package main

import (
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/finfreezer/httpfromtcp/internal/request"
	"github.com/finfreezer/httpfromtcp/internal/response"
	"github.com/finfreezer/httpfromtcp/internal/server"
)

const port = 42069

func main() {
	handler := func(w io.Writer, req *request.Request) *server.HandlerError {
		newErr := server.HandlerError{}
		if strings.Contains(req.RequestLine.RequestTarget, "/yourproblem") {
			newErr = server.HandlerError{StatusCode: response.BadRequest, Message: "Your problem is not my problem\n"}
		} else if strings.Contains(req.RequestLine.RequestTarget, "/myproblem") {
			newErr = server.HandlerError{StatusCode: response.InternalServerError, Message: "Woopsie, my bad\n"}
		} else {
			newErr = server.HandlerError{StatusCode: response.OK, Message: "All good, frfr\n"}
		}
		return &newErr
	}
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
