package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"
)

type status int

const (
	ok   status = 0
	fail status = 1
)

var files, output, packagename string

func main() {
	flag.StringVar(&files, "files", "", "Comma separated list of"+
		"files to make code for")
	flag.StringVar(&output, "output", "", "Output file name")
	flag.StringVar(&packagename, "package", "", "Name of output package.")
	flag.Parse()

	fl, err := os.OpenFile(output, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		panic(err)
	}

	defer fl.Close()

	for inf := range run(fl) {
		var err error
		switch v := inf.(type) {
		case error:
			_, err = fmt.Fprintf(os.Stderr, "Error: %s\n", v.Error())

		case string:
			_, err = os.Stdout.Write([]byte(v + "\n"))
		default:
			_, err = fmt.Fprint(os.Stdout, v)
		}

		if err != nil {
			panic(fmt.Sprintf("Fatal error: %s\n", err.Error()))
		}
	}

}

func run(w io.Writer) (chn <-chan interface{}) {
	c := make(chan interface{})
	chn = c
	go func() {
		defer close(c)
		if output == "" {
			c <- fmt.Errorf("No output file provided!")
			return
		}

		_, err := fmt.Fprintf(w,
			`package %s

import "github.com/TShadwell/fweight/resource"
import "time"

func resourceWillTime(s string) (t time.Time) { t, _ = time.Parse(time.RFC3339, s); return }
`, packagename)

		if err != nil {
			c <- err
			return
		}

		ns := time.Now().Format(time.RFC3339)

		//split
		for _, g := range strings.Split(files, ",") {
			bt, err := ioutil.ReadFile(g)
			if err != nil {
				c <- err
				return
			}

			fmt.Fprintf(
				w,
				"\nvar %s = resource.Resource{%+q, %+q, resourceWillTime(%+q)}",
				strings.Replace(path.Base(g), ".", "_", -1),
				path.Base(g),

				bt,
				ns,
			)

		}
		return
	}()
	return
}
