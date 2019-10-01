package log_test

import (
	"github.com/anthonycorbacho/foundation/log"
)

func Example_basic() {
	logger, err := log.New()
	if err != nil {
		panic(err)
	}
	defer logger.Close()

	logger.Info("This is my log", log.String("Key", "String value"))
}
