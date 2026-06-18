package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
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

		if strings.Contains(req.RequestLine.RequestTarget, "/httpbin") {
			proxyTarget := "https://httpbin.org" +
				strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin")

			err := helperHttpProxy(w, proxyTarget)
			if err != nil {
				log.Println(err)
				return
			}
			return
		}

		if strings.Contains(req.RequestLine.RequestTarget, "/video") {
			err := helperGetVideo(w)
			if err != nil {
				log.Println(err)
				return
			}
			return
		}

		err := helperNoProblem(w)
		if err != nil {
			log.Println(err)
			return
		}
	} //End of handler

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

func helperGetVideo(w *response.Writer) error {
	err := response.WriteStatusLine(w.ContentWriter, response.OK)
	if err != nil {
		log.Println(err)
		return err
	}

	fullBodyBuf := []byte{}
	fmt.Println("Writing data chunks...")
	fullBodyBuf, err = os.ReadFile("assets/vim.mp4")
	if err != nil {
		log.Println(err)
		return err
	}

	h := response.GetDefaultHeaders(len(fullBodyBuf))
	h = response.SetDefaultHeaders(h, []string{"content-type"}, []string{"video/mp4"})
	err = w.WriteHeaders(h)
	if err != nil {
		log.Println(err)
		return err
	}

	fmt.Println("Finished writing body.")
	n, err := w.ContentWriter.Write(fullBodyBuf)
	if err != nil {
		log.Println(err)
		return err
	}
	lenInt, err := strconv.Atoi(h["content-length"])
	if n != lenInt {
		log.Println("Error writing full data.")
	}

	return nil
}

func helperHttpProxy(w *response.Writer, proxyPath string) error {
	withTrailers := false

	if strings.Contains(proxyPath, "/html") {
		withTrailers = true
	}
	newResp, RespErr := http.Get(proxyPath)
	log.Println(newResp.StatusCode)

	if RespErr != nil {
		err := response.WriteStatusLine(w.ContentWriter, response.BadRequest)
		if err != nil {
			log.Println(err)
			return err
		}

		errorBody := []byte("Bad GET request.")
		h := response.GetDefaultHeaders(len(errorBody))
		err = w.WriteHeaders(h)
		if err != nil {
			log.Println(err)
			return err
		}

		_, err = w.ContentWriter.Write(errorBody)
		if err != nil {
			log.Println(err)
			return err
		}

		log.Println(RespErr)
		return RespErr
	}

	defer newResp.Body.Close()
	err := response.WriteStatusLine(w.ContentWriter, response.OK)
	if err != nil {
		log.Println(err)
		return err
	}

	h := response.GetDefaultHeaders(0)
	h = response.RemoveSetHeader(h, "content-length")
	h["transfer-encoding"] = "chunked"
	if withTrailers {
		h["trailer"] = "x-content-sha256, x-content-length"
	}
	err = w.WriteHeaders(h)
	if err != nil {
		log.Println(err)
		return err
	}

	chunkBuf := make([]byte, 32)
	fullBodyBuf := []byte{}
	bodySize := 0
	fmt.Println("Writing body chunks...")
	for {
		n, err := newResp.Body.Read(chunkBuf)
		fullBodyBuf = append(fullBodyBuf, chunkBuf[:n]...)

		if n > 0 {
			m, err := w.WriteChunkedBody(chunkBuf[:n])
			bodySize += m
			if err != nil {
				log.Println(err)
				return err
			}
		}
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}
	}
	if withTrailers {
		fmt.Println("Adding checksum...")
		sum := sha256.Sum256(fullBodyBuf)
		hashStr := fmt.Sprintf("%x", sum)
		h["x-content-sha256"] = hashStr
		fmt.Println(hashStr)
		contentLenStr := strconv.Itoa(bodySize)
		h["x-content-length"] = contentLenStr
	}

	w.Headers = h
	_, chunkErr := w.WriteChunkedBodyDone(withTrailers)
	fmt.Println("Finished writing body.")
	fmt.Println(string(fullBodyBuf))
	fmt.Println(bodySize)
	if chunkErr != nil {
		log.Println(chunkErr)
		return chunkErr
	}

	return nil
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
