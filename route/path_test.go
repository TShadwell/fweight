package route

import (
	"fmt"
)

func ExampleAmpersand() {
	name := Ampersand("/user/bob/images/", "/user/", "/")
	fmt.Println(name)

	name = Ampersand("/user/anne.html", "/user/", "./")
	fmt.Println(name)

	name = Ampersand("/user/anne/images/cat.png", "/user/", "./")
	fmt.Println(name)
	// Output:
	// bob
	// anne
	// anne
}

func ExamplePartN() {
	name := PartN("/user/bob/images", 1)
	fmt.Println(name)

	annepage := PartN("/user/anne.html", 1)
	fmt.Println(annepage)

	imagename := PartN("user/anne/imags/cat.png", 3)
	fmt.Println(imagename)

	empty := PartN("/user/anne/images/", 3)
	fmt.Println(empty)

	// Output:
	// bob
	// anne.html
	// cat.png
	//
}

func ExampleTrimpast() {
	const fn = "log.txt.gz"

	//Prints the filename, minus the extension.
	fmt.Println(TrimPast(fn, "."))

	const path = "bin/pattern"

	//prints the first segement of the path
	fmt.Println(TrimPast(path, "/"))

	// Output:
	// log
	// bin
}
