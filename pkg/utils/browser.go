package utils

import (
	"fmt"

	"github.com/skratchdot/open-golang/open"
)

func TryOpenBrowser(url string) {
	if err := open.Run(url); err != nil {
		fmt.Printf("Browse to: %s", url)
	}
}
