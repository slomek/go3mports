package example

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/slomek/go3mports/example/performers"
)

func Welcome() {
	names, err := performers.Get(4)
	if err != nil {
		fmt.Printf("Sorry, we can't find performers: %v\n", errors.Cause(err))
	}
	performers := strings.Join(names, ", ")
	line := fmt.Sprintf("Tonight performing are: %s", performers)
	fmt.Println(line)
}
