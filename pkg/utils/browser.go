package utils

import (
	"fmt"

	"github.com/skratchdot/open-golang/open"
)

func TryOpenBrowser(message string, url string) {
	fmt.Printf("%s %s\n", message, url)
	open.Run(url)
}
