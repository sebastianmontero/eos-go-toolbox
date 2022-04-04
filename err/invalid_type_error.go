package err

import "fmt"

type InvalidTypeError struct {
	Label        string
	ExpectedType string
	Value        interface{}
}

func (c *InvalidTypeError) Error() string {
	return fmt.Sprintf("%v. Expected Type: %v", c.Label, c.ExpectedType)
}
