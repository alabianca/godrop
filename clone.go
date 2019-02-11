package godrop

import (
	"io"
	"log"
	"os"
)

type Clone struct {
	sesh *Session
	// 2 buffered channels to prevent dead locks as they depend on each other
	readHeader       chan int
	readContent      chan Header
	transferComplete chan int
}

// CloneDir clones a directory shared by a godrop peer
func (c *Clone) CloneDir() {

	for {
		select {
		case <-c.readHeader:
			//read header
			if err := c.readheader(); err != nil {
				log.Println(err)
			}
		case header := <-c.readContent:
			//read content
			if err := c.readcontent(header); err != nil {
				log.Println(err)
			}

		case <-c.transferComplete:
			return
		}
	}
}

func (c *Clone) readheader() error {
	header, err := c.sesh.ReadHeader()

	if header.Name != "" {
		c.readContent <- header
	}

	if header.IsComplete() {
		c.transferComplete <- 1
	}

	return err
}

func (c *Clone) readcontent(h Header) error {

	// content is a file
	if !h.IsDir() {
		if err := c.readFile(h); err != nil {
			return err
		}

		c.readHeader <- 1
		return nil
	}

	// content is a directory. create it
	if err := os.Mkdir(h.Path, 0700); err != nil {
		return err
	}

	c.readHeader <- 1
	return nil
}

func (c *Clone) readFile(h Header) error {
	file, err := os.Create(h.Path)
	defer file.Close()
	if err != nil {
		return err
	}

	var receivedByts int64

	for {
		if (h.Size - receivedByts) < BUF_SIZE {
			io.CopyN(file, c.sesh, (h.Size - receivedByts))
			break
		}

		io.CopyN(file, c.sesh, BUF_SIZE)
		receivedByts += BUF_SIZE
	}

	return nil
}
