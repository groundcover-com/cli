package utils

import (
	"fmt"

	"github.com/skratchdot/open-golang/open"
)

func TryOpenBrowser(url string) {
	fmt.Printf("Browse to: %s\n", url)
	open.Run(url)
}
