package main

import (
	"bufio"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"os"
)

func main() {
	flag.Parse()
	files := flag.Args()
	bufWriter := bufio.NewWriter(os.Stdout)
	if len(files) > 0 {
		for _, file := range files {
			f, err := os.OpenFile(file, os.O_RDONLY, 0744)
			if err != nil {
				fmt.Errorf("Couldn't open file: %s\n", f)
				os.Exit(-1)
			}
			bufReader := bufio.NewReader(f)
			printGzip(bufReader, bufWriter)
			bufWriter.Flush()
			f.Close()
		}
	} else {
		bufReader := bufio.NewReader(os.Stdin)
		printGzip(bufReader, bufWriter)
		bufWriter.Flush()
	}
}

func printGzip(in io.Reader, out io.Writer) {
	gzReader, err := gzip.NewReader(in)
	if err != nil {
		panic(err)
	}
	defer gzReader.Close()

	p := make([]byte, 4096)
	for {
		n, err := gzReader.Read(p)
		if n > 0 {
			out.Write(p[0:n])
		}
		if err != nil {
			break
		}
	}
}
