// Package version provides a simple logic that will inject application version at build time.
// To use this package, you need to setup your Makefile to inject data via `LDFLAGS`.
//
//		VERSION_PACKAGE	= github.com/anthonycorbacho/foundation/version
//
//		GO_LDFLAGS 	:="
//		GO_LDFLAGS	+= -X $(VERSION_PACKAGE).version=$(VERSION)
//		GO_LDFLAGS 	+= -X $(VERSION_PACKAGE).buildDate=$(shell date +'%Y-%m-%dT%H:%M:%SZ')
//		GO_LDFLAGS 	+= -X $(VERSION_PACKAGE).gitCommit=$(shell git rev-parse HEAD)
//		GO_LDFLAGS 	+= -X $(VERSION_PACKAGE).gitTreeState=$(if $(shell git status --porcelain),dirty,clean)
//		GO_LDFLAGS 	+="
//
// When calling go build, add the ld flags `-ldflags $(GO_LDFLAGS)`.
// You can then get the application version by doing
//
//		v := version.Get()
//		logger.Info("starting auth service",
//			zap.String("version", v.Version),
//			zap.String("build date", v.BuildDate),
//			zap.String("git commit", v.GitCommit),
//			zap.String("git tree state", v.GitTreeState))
//
package version
