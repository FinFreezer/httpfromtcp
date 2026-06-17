package main

import (
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

	handler := func(w *response.Writer, req *request.Request) {
		if strings.Contains(req.RequestLine.RequestTarget, "/yourproblem") {
			err := helperYourProblem(w)
			if err != nil {
				log.Println(err)
				return
			}
		}

		if strings.Contains(req.RequestLine.RequestTarget, "/myproblem") {
			err := helperMyProblem(w)
			if err != nil {
				log.Println(err)
				return
			}
		}

		err := helperNoProblem(w)
		if err != nil {
			log.Println(err)
			return
		}
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

func helperYourProblem(w *response.Writer) error {
	err := response.WriteStatusLine(w.ContentWriter, response.BadRequest)
	if err != nil {
		log.Println(err)
		return err
	}
	err = w.WriteStatusLine(response.BadRequest)
	if err != nil {
		log.Println(err)
		return err
	}

	_, err = w.WriteBody([]byte("Your request honestly kinda sucked."))
	if err != nil {
		log.Println(err)
		return err
	}

	err = w.ReplaceHTMLHeader([]byte("Bad Request"))
	if err != nil {
		log.Println(err)
		return err
	}

	h := response.GetDefaultHeaders(len(w.ResponseHTML))
	h = response.SetDefaultHeaders(h, []string{"content-type"}, []string{"text/html"})
	err = w.WriteHeaders(h)
	if err != nil {
		log.Println(err)
		return err
	}

	_, err = w.ContentWriter.Write([]byte(w.ResponseHTML))
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func helperMyProblem(w *response.Writer) error {
	err := response.WriteStatusLine(w.ContentWriter, response.InternalServerError)
	if err != nil {
		log.Println(err)
		return err
	}

	err = w.WriteStatusLine(response.InternalServerError)
	if err != nil {
		log.Println(err)
		return err
	}

	_, err = w.WriteBody([]byte("Okay, you know what? This one is on me."))
	if err != nil {
		log.Println(err)
		return err
	}

	err = w.ReplaceHTMLHeader([]byte("Internal Server Error"))
	if err != nil {
		log.Println(err)
		return err
	}

	h := response.GetDefaultHeaders(len(w.ResponseHTML))
	h = response.SetDefaultHeaders(h, []string{"content-type"}, []string{"text/html"})
	err = w.WriteHeaders(h)
	if err != nil {
		log.Println(err)
		return err
	}

	_, err = w.ContentWriter.Write([]byte(w.ResponseHTML))
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func helperNoProblem(w *response.Writer) error {
	err := response.WriteStatusLine(w.ContentWriter, response.OK)
	if err != nil {
		log.Println(err)
		return err
	}

	err = w.WriteStatusLine(response.OK)
	if err != nil {
		log.Println(err)
		return err
	}

	_, err = w.WriteBody([]byte("Your request was an absolute banger."))
	if err != nil {
		log.Println(err)
		return err
	}

	err = w.ReplaceHTMLHeader([]byte("Success!"))
	if err != nil {
		log.Println(err)
		return err
	}

	h := response.GetDefaultHeaders(len(w.ResponseHTML))
	h = response.SetDefaultHeaders(h, []string{"content-type"}, []string{"text/html"})
	err = w.WriteHeaders(h)
	if err != nil {
		log.Println(err)
		return err
	}

	_, err = w.ContentWriter.Write([]byte(w.ResponseHTML))
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}
