package main

import (
	"flag"
	"fmt"
	"github.com/TShadwell/fweight/route"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

type status int

const (
	ok   status = 0
	fail status = 1
)

var files, output, packagename string

func main() {
	fmt.Println(":¬)")
	flag.StringVar(&files, "files", "", "Comma separated list of" +
		"files to make code for")
	flag.StringVar(&output, "output", "", "Output file name")
	flag.StringVar(&packagename, "package", "", "Name of output package.")
	flag.Parse()

	fl, err := os.Open(output)
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
	c <- ":¬)"
	go func() {
		defer close(c)
		c <- ":¬)"
		if output == "" {
			c <- fmt.Errorf("No output file provided!")
			return
		}

		_, err := fmt.Fprintf(w, "package %s\nimport \"github.com/TShadwell/fweight/resource\"\n", packagename)
		if err != nil {
			c <- err
			return
		}

		//split
		for _, g := range strings.Split(files, ",") {
			bt, err := ioutil.ReadFile(g)
			if err != nil {
				c <- err
				return
			}

			fmt.Fprintf(
				w,
				"\nvar %s = resource.Resource{%+q, %+q}",
				route.TrimPast(
					path.Base(g),
					".",
				),
				g,
				bt,
			)


		}
		return
	}()
	return
}
