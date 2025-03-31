package wrapper

import "fmt"

func WrapError(base error, err error) error {
	return fmt.Errorf("%w: %w", base, err)
}

