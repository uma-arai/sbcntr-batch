package utils

import (
	"fmt"
	"runtime/debug"
)

// GetStackWithError は、エラーとスタックトレースを組み合わせて返します
func GetStackWithError(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%w\nStack trace:\n%s", err, debug.Stack())
}
