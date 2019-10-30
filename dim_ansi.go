package main

import (
	"bytes"
	"github.com/creack/pty"
	"golang.org/x/crypto/ssh/terminal"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

func main() {
	c := exec.Command(os.Args[1], os.Args[2:]...)

	ptmx, err := pty.Start(c)
	if err != nil {
		panic(err)
	}

	defer func() { _ = ptmx.Close() }()

	// pty size.
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)
	go func() {
		for range ch {
			if err := pty.InheritSize(os.Stdin, ptmx); err != nil {
				log.Printf("error resizing pty: %s", err)
			}
		}
	}()
	ch <- syscall.SIGWINCH // Initial resize.

	// Set stdin in raw mode.
	oldState, err := terminal.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	defer func() { _ = terminal.Restore(int(os.Stdin.Fd()), oldState) }()

	// Streams
	go func() { _, _ = io.Copy(ptmx, os.Stdin) }()
	go func() { copyStream(ptmx, os.Stderr) }()
	copyStream(os.Stdout, ptmx)
}

func copyStream(dest io.Writer, source io.Reader) {
	buf := make([]byte, 16*1024)
	for {
		n, err := source.Read(buf)
		if err != nil {
			break
		}

		if n > 0 {
			bufslice := buf[0:n]
			bufslice = removeSeq("\033[1;", bufslice)
			bufslice = removeSeq("\033[01;", bufslice)
			bufslice = removeSeq("\033[37", bufslice)

			w, err := dest.Write(bufslice)
			if err != nil {
				break
			}

			if w != len(bufslice) {
				log.Println("could not write all data")
				break
			}
		}
	}
}

func removeSeq(seq string, buf []byte) []byte {
	if bytes.Contains(buf, []byte(seq)) {
		return bytes.ReplaceAll(buf, []byte(seq), []byte("\033["))
	}

	return buf
}
