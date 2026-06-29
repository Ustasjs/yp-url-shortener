package notmain

import "os"

// main here belongs to a non-main package, so os.Exit must be ignored.
func main() {
	os.Exit(1)
}
