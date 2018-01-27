package performers

import (
	"errors"

	"strings"
)

func Get(n int) ([]string, error) {
	if n > 4 {
		return nil, errors.New("there are enough performers")
	}
	return []string{"Wayne", "Jeff", "Colin", strings.ToTitle("ryan")}, nil
}
