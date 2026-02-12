package assertions

import (
	"errors"
	"raidline/ripple/pgk/logger"
)

func NonError(err error, message string) {
	if err != nil {
		logger.Error("Error : [%s], message: [%s]", err.Error(), message)
		panic(err)
	}
}

func NotNil(obj any, message string) {
	if obj == nil {
		err := errors.New("Obj is null and should not")
		logger.Error("Error : [%s], message : [%s]", err.Error(), message)
		panic(err)
	}
}

func Condition(condition bool, message string) {
	if !condition {
		err := errors.New("Condition could not be verified")
		logger.Error("Error : [%s], message : [%s]", err.Error(), message)
		panic(err)
	}
}
