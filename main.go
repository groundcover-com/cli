package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/getsentry/sentry-go"
	"groundcover.com/cmd"
	sentry_utils "groundcover.com/pkg/sentry"
	"groundcover.com/pkg/ui"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

func main() {
	var err error

	klogFlagSet := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	klog.InitFlags(klogFlagSet)
	klogFlagSet.Set("logtostderr", "false")

	rest.SetDefaultWarningHandler(rest.NoWarnings{})

	environment := "prod"
	release := fmt.Sprintf("cli@%s", cmd.BinaryVersion)

	if cmd.IsDevVersion() {
		environment = "dev"
	}

	sentryClientOptions := sentry_utils.GetSentryClientOptions(environment, release)
	if err = sentry.Init(sentryClientOptions); err != nil {
		ui.SingletonWriter.PrintErrorMessageln(err.Error())
		panic(err)
	}
	defer sentry.Flush(sentry_utils.FLUSH_TIMEOUT)

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
