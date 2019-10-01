// Package log provides a application log that formalize a log format.
//
//		logger, err := log.New()
//		if err != nil {
//			panic(err)
//		}
//		defer logger.Close()
//
//		logger.Info("This is my log", log.String("Key", "String value"))
//
package log
