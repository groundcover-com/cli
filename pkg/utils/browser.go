package utils

import (
	"fmt"

	"github.com/skratchdot/open-golang/open"
)

func OpenBrowser(url string) {
	if err := open.Run(url); err != nil {
		fmt.Printf("You can browse to: %s", url)
	}
}
