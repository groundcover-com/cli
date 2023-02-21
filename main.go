package main

import (
	"context"
	"flag"
	"os"
	"os/signal"

	"github.com/getsentry/sentry-go"
	"groundcover.com/cmd"
	"groundcover.com/pkg/segment"
	sentry_utils "groundcover.com/pkg/sentry"
	"groundcover.com/pkg/ui"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

const APP_NAME = "cli"

func main() {
	var err error

	klogFlagSet := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	klog.InitFlags(klogFlagSet)
	klogFlagSet.Set("logtostderr", "false")

	rest.SetDefaultWarningHandler(rest.NoWarnings{})

	environment := "prod"
	if cmd.IsDevVersion() {
		environment = "dev"
	}

	sentryClientOptions := sentry_utils.GetSentryClientOptions(APP_NAME, environment, cmd.BinaryVersion)
	if err = sentry.Init(sentryClientOptions); err != nil {
		ui.GlobalWriter.PrintErrorMessageln(err.Error())
		panic(err)
	}
	defer sentry.Flush(sentry_utils.FLUSH_TIMEOUT)

	segmentConfig := segment.GetConfig(APP_NAME, cmd.BinaryVersion)
	if err = segment.Init(segmentConfig); err != nil {
		ui.GlobalWriter.PrintErrorMessageln(err.Error())
		panic(err)
	}
	defer segment.Close()

	ctx, cleanup := contextWithSignalInterrupt()
	defer cleanup()

	cmd.ExecuteContext(ctx)
}

func contextWithSignalInterrupt() (context.Context, func()) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	ctx, cancel := context.WithCancel(context.Background())

	cleanup := func() {
		signal.Stop(signalChan)
		cancel()
	}

	go func() {
		select {
		case <-signalChan:
			cancel()
		case <-ctx.Done():
		}
	}()

	return ctx, cleanup
}
