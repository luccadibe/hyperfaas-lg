package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/bojand/ghz/printer"
	"github.com/bojand/ghz/runner"
)

type config struct {
	host  string
	proto string
	call  string
}

func main() {

	var cfg config

	flag.StringVar(&cfg.host, "host", "localhost:8080", "host to run the benchmark against")
	flag.StringVar(&cfg.proto, "proto", "", "proto file to use")
	flag.StringVar(&cfg.call, "call", "", "call to run")
	flag.Parse()

	report, err := runner.Run(cfg.call, cfg.host,
		runner.WithProtoFile(cfg.proto, []string{"proto/"}),
		runner.WithConfigFromFile("config.json"),
		runner.WithInsecure(true),
		runner.WithConcurrencySchedule("line"),
		runner.WithConcurrencyStart(1),
		runner.WithConcurrencyEnd(10),
		runner.WithConcurrencyStep(1),
		runner.WithTimeout(10*time.Second),
		runner.WithAsync(false),
		runner.WithRunDuration(1*time.Minute),
	)
	if err != nil {
		log.Fatal(err)
	}

	csvFile, err := os.Create("out.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer csvFile.Close()

	printer := printer.ReportPrinter{
		Report: report,
		Out:    csvFile,
	}
	printer.Print("csv")
}
