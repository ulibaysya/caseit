package user

import "fmt"

type ID int64

func (i ID) Validate() error {
	if i < 0 {
		return fmt.Errorf("less than zero: %v", i)
	} else if i == 0 {
		return fmt.Errorf("equal to zero: %v", i)
	}
	return nil
}
