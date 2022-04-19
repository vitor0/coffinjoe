package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/genuinetools/pkg/cli"
)

var (
	username string
	password string
	date     string
	output   string
	timeout  int
)

const host = "https://selo.tjsc.jus.br/selo/CertidaoService" //"https://selo.tjsc.jus.br/selo_teste/CertidaoService" --> HML

func main() {
	p := cli.NewProgram()
	p.Description = "API client to consume the death certificate from tjsc"
	p.FlagSet = flag.NewFlagSet("global", flag.ExitOnError)

	p.GitCommit = GITCOMMIT
	p.Version = VERSION

	//p.FlagSet.StringVar(&username, "username", "convenio", "TJ-SC username")
	//p.FlagSet.StringVar(&username, "u", "convenio", "TJ-SC username")
	p.FlagSet.StringVar(&username, "username", "convenio_cge", "TJ-SC username")
	p.FlagSet.StringVar(&username, "u", "convenio_cge", "TJ-SC username")

	//	p.FlagSet.StringVar(&password, "password", "selodigital", "TJ-SC password")
	//	p.FlagSet.StringVar(&password, "p", "selodigital", "TJ-SC password")
	p.FlagSet.StringVar(&password, "password", "myp1y2MOds", "TJ-SC password")
	p.FlagSet.StringVar(&password, "p", "myp1y2MOds", "TJ-SC password")

	p.FlagSet.StringVar(&date, "date", "2018-06-04", "specific date of request (YYYY-MM-DD)")
	p.FlagSet.StringVar(&date, "d", "2018-06-04", "specific date of request (YYYY-MM-DD)")

	p.FlagSet.StringVar(&output, "output", "certificates", "Output certificates JSON file")
	p.FlagSet.StringVar(&output, "o", "certificates", "Output certificates JSON file")

	p.FlagSet.IntVar(&timeout, "timeout", 20, "Http client timeout (in seconds)")
	p.FlagSet.IntVar(&timeout, "t", 20, "Http client timeout (in seconds)")

	p.Before = func(ctx context.Context) error {
		signals := make(chan os.Signal)
		signal.Notify(signals, os.Interrupt)
		signal.Notify(signals, syscall.SIGTERM)
		_, cancel := context.WithCancel(ctx)
		go func() {
			for sig := range signals {
				cancel()
				log.Printf("Received %s, exiting", sig.String())
				os.Exit(0)
			}
		}()
		return nil
	}
	p.Action = func(ctx context.Context, args []string) error {
		options := []Option{
			WithCredentials(username, password),
			WithTimeout(timeout),
		}
		client := NewClient(host, options...)
		cd, err := client.GetDeatchCertificateByDate(ctx, date)
		if err != nil {
			return err
		}
		if err := exportJSON(cd, output); err != nil {
			return err
		}
		return exportXML(cd, output)
	}
	p.Run()
}
