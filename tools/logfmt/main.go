package main

import (
	"bufio"
	"bytes"
	"encoding/json"
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
			formatJson(bufReader, bufWriter)
			bufWriter.Flush()
			f.Close()
		}
	} else {
		bufReader := bufio.NewReader(os.Stdin)
		formatJson(bufReader, bufWriter)
		bufWriter.Flush()
	}
}

func formatJson(in *bufio.Reader, out io.Writer) {
	buf := new(bytes.Buffer)
	for {
		line, err := in.ReadBytes('\n')
		if err := json.Indent(buf, line, "", "  "); err != nil {
			break
		}
		out.Write(buf.Bytes())
		buf.Reset()
		if err != nil {
			break
		}
	}
}
