package srvcerrors

import "fmt"

var (
	ErrNotFound           = fmt.Errorf("error not found")
	ErrOrderAlreadyExists = fmt.Errorf("order already exists")
	ErrDatabase           = fmt.Errorf("database error")
	ErrInvalidInput       = fmt.Errorf("invalid input")
	ErrKafka              = fmt.Errorf("kafka error")
)
