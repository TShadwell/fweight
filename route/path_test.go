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
