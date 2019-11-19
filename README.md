# Foundation [![CircleCI](https://circleci.com/gh/anthonycorbacho/foundation/tree/master.svg?style=svg)](https://circleci.com/gh/anthonycorbacho/foundation/tree/master)

This starter kit is a starting point for building production grade scalable Go service applications.
The goal of this Foundation is to provide a proven starting point for new projects that reduce the repetitive tasks in getting a new project launched to production.

This project should not be considered as a "framework".
This project leaves you in control of your project’s architecture and development.

## How to use

```go
// create a GRPC server from github.com/anthonycorbacho/foundation/grpc import.
srv := grpc.NewServer()
defer srv.Stop()
pb.RegisterXXXXXXServer(srv, &myStruct{})

// Create foundation service with prometheus and jeager exporter.
svc := foundation.NewService(":8100", foundation.Name("service_name"))
svc.WithPrometheusExporter(":8101")
svc.WithJaegerExporter("http://127.0.0.1:14268/api/traces", trace.AlwaysSample())

// start the service.
if err := svc.Serve(srv); err != nil {
	logger.Fatal("failed to serve", log.Error(err))
}
```

## Contributing
We are welcoming any contribution to the project. If you have a use case or pattern you want to include please feel free to open an issue or create a Pull Request.
