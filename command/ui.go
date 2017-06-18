package command

import "fmt"

// Abort print an error and return 255.
func Abort(format string, a ...interface{}) int {
	fmt.Printf("abort: "+format, a...)
	return 255
}
