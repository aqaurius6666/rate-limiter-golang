package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/sonntuet1997/medical-chain-utils/common_service"
	pb2 "github.com/sonntuet1997/medical-chain-utils/common_service/pb"

	"cloud.google.com/go/profiler"
	"contrib.go.opencensus.io/exporter/stackdriver"
	"github.com/urfave/cli/v2"
	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
	"google.golang.org/grpc"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func runMain(appCtx *cli.Context) error {
	var wg sync.WaitGroup
	ctx, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()
	server, err := InitializeLimiterServiceServer(ctx, logger, RedisOpts{
		Uri:      RedisURI(appCtx.String("redis-uri")),
		Pass:     RedisPassword(appCtx.String("redis-pass")),
		Username: RedisUsername(appCtx.String("redis-username")),
		Remote:   RemoteURL(appCtx.String("remote-url")),
	})
	if err != nil {
		return err
	}
	defer func(service *RedisService) {
		err := service.Close()
		if err != nil {
			panic("cannot shutdown redis database")
		}
	}(server.Service)

	if appCtx.Bool("disable-tracing") {
		logger.Info("Tracing disabled.")
	} else {
		logger.Info("Tracing enabled.")
		go initTracing()
	}
	if appCtx.Bool("disable-profiler") {
		logger.Info("Profiling disabled.")
	} else {
		logger.Info("Profiling enabled.")
		go initProfiling(serviceName, appCtx.String("runtime-version"))
	}
	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%d", appCtx.Int("grpc-port")))
	if err != nil {
		return err
	}
	defer func() { _ = grpcListener.Close() }()
	var srv *grpc.Server
	commonServer := common_service.NewCommonServiceServer(logger, appCtx.Bool("allow-kill"))
	wg.Add(1)
	go func() {
		defer wg.Done()
		if appCtx.Bool("disable-stats") {
			logger.Info("Stats disabled.")
			srv = grpc.NewServer()
		} else {
			logger.Info("Stats enabled.")
			srv = grpc.NewServer(grpc.StatsHandler(&ocgrpc.ServerHandler{}))
		}
		healthpb.RegisterHealthServer(srv, commonServer)
		pb2.RegisterCommonServiceServer(srv, commonServer)
		reflection.Register(srv)
		logger.WithField("port", appCtx.Int("grpc-port")).Info("listening for gRPC connections")
		// _ = srv.Serve(grpcListener)
		if err := srv.Serve(grpcListener); err != nil {
			logger.Fatalf("failed to serve: %v", err)
		}
	}()

	proxySrv := &http.Server{
		Addr:    fmt.Sprintf(":%d", appCtx.Int("http-port")),
		Handler: server.G,
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err != nil {
			logger.Errorf("err: %v", err)
		}
		server.RegisterProxy()
		if err := proxySrv.ListenAndServe(); err != nil {
			logger.Printf("listen: %s\n", err)
		}
	}()

	// Start pprof server
	pprofListener, err := net.Listen("tcp", fmt.Sprintf(":%d", appCtx.Int("pprof-port")))
	if err != nil {
		return err
	}
	defer func() { _ = pprofListener.Close() }()

	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.WithField("port", appCtx.Int("pprof-port")).Info("listening for pprof requests")
		sSrv := new(http.Server)
		_ = sSrv.Serve(pprofListener)
	}()
	// Start signal watcher
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGKILL)
		select {
		case s := <-sigCh:
			logger.WithField("signal", s.String()).Infof("shutting down due to signal")
			srv.Stop()
			_ = grpcListener.Close()
			_ = pprofListener.Close()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := proxySrv.Shutdown(ctx); err != nil {
				logger.Fatal("Server Shutdown:", err)
			}
			cancelFn()
			logger.WithField("signal", s.String()).Infof("shutdown success!")
		case <-ctx.Done():
		}
	}()
	// Keep running until we receive a signal
	wg.Wait()
	return nil
}

func initTracing() {
	// initJaegerTracing()
	initStackdriverTracing()
}

// func initJaegerTracing() {
// 	svcAddr := os.Getenv("JAEGER_SERVICE_ADDR")
// 	if svcAddr == "" {
// 		logger.Info("jaeger initialization disabled.")
// 		return
// 	}

// 	// Register the Jaeger exporter to be able to retrieve
// 	// the collected spans.
// 	exporter, err := jaeger.NewExporter(jaeger.Options{
// 		CollectorEndpoint: fmt.Sprintf("http://%s", svcAddr),
// 		Process: jaeger.Process{
// 			ServiceName: serviceName,
// 		},
// 	})
// 	if err != nil {
// 		logger.Fatal(err)
// 	}
// 	trace.RegisterExporter(exporter)
// 	logger.Info("jaeger initialization completed.")
// }

func initStats(exporter *stackdriver.Exporter) {
	view.SetReportingPeriod(60 * time.Second)
	view.RegisterExporter(exporter)
	if err := view.Register(ocgrpc.DefaultServerViews...); err != nil {
		logger.Warn("Error registering default server views")
	} else {
		logger.Info("Registered default server views")
	}
}

func initStackdriverTracing() {
	// TODO(ahmetb) this method is duplicated in other microservices using Go
	// since they are not sharing packages.
	for i := 1; i <= 3; i++ {
		exporter, err := stackdriver.NewExporter(stackdriver.Options{})
		if err != nil {
			logger.Infof("failed to initialize stackdriver exporter: %+v", err)
		} else {
			trace.RegisterExporter(exporter)
			logger.Info("registered Stackdriver tracing")

			// Register the views to collect server stats.
			initStats(exporter)
			return
		}
		d := time.Second * 10 * time.Duration(i)
		logger.Infof("sleeping %v to retry initializing Stackdriver exporter", d)
		time.Sleep(d)
	}
	logger.Warn("could not initialize Stackdriver exporter after retrying, giving up")
}

func initProfiling(service, version string) {
	// TODO(ahmetb) this method is duplicated in other microservices using Go
	// since they are not sharing packages.
	for i := 1; i <= 3; i++ {
		if err := profiler.Start(profiler.Config{
			Service:        service,
			ServiceVersion: version,
			// ProjectID must be set if not running on GCP.
			// ProjectID: "my-project",
		}); err != nil {
			logger.Warnf("failed to start profiler: %+v", err)
		} else {
			logger.Info("started Stackdriver profiler")
			return
		}
		d := time.Second * 10 * time.Duration(i)
		logger.Infof("sleeping %v to retry initializing Stackdriver profiler", d)
		time.Sleep(d)
	}
	logger.Warn("could not initialize Stackdriver profiler after retrying, giving up")
}
