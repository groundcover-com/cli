package utils

import (
	"fmt"

	"github.com/skratchdot/open-golang/open"
)

func TryOpenBrowser(url string) {
	open.Run(url)
	fmt.Printf("Browse to: %s\n", url)
}
