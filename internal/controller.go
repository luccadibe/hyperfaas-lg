package internal

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/3s-rg-codes/HyperFaaS/proto/common"
	"gopkg.in/yaml.v2"
)

type Controller struct {
	Config    *Config
	collector *Collector
	funcMgr   *FunctionManager
	l         *slog.Logger
}

type Config struct {
	GenerateWorkload bool                     `yaml:"generate_workload"`
	LeafAddress      string                   `yaml:"leaf_address"`
	Seed             int64                    `yaml:"seed,omitempty"`
	MaxDuration      time.Duration            `yaml:"max_duration"`
	Timeout          int32                    `yaml:"timeout"`
	Patterns         map[string]*PhasePattern `yaml:"patterns"`
	Workload         *Workload                `yaml:"workload,omitempty"`
}

type Workload struct {
	LeafAddress string        `yaml:"leaf_address"`
	MaxDuration time.Duration `yaml:"max_duration"`
	Timeout     int32         `yaml:"timeout"`
	Phases      []TestPhase   `yaml:"phases"` // should be ordered by start time ascending
}

type TestPhase struct {
	Name       string        `yaml:"name"`
	Type       string        `yaml:"type"`       // "constant" | "variable"
	StartTime  time.Duration `yaml:"start_time"` // Relative to workload start
	Duration   time.Duration `yaml:"duration"`
	StartRPS   int           `yaml:"start_rps"`
	EndRPS     int           `yaml:"end_rps,omitempty"`
	Step       int           `yaml:"step,omitempty"` // For ramping increment/decrement
	ImageTag   string        `yaml:"image_tag"`
	FunctionID string        // function target
}

func (c *Controller) Run() {

	client := NewLeafClient(c.Config.LeafAddress)
	fmt.Println("Creating functions")
	c.CreateFunctions()

	fmt.Println("Starting workload, max duration:", c.Config.MaxDuration)
	ctx, cancel := context.WithTimeout(context.Background(), c.Config.MaxDuration)
	defer cancel()

	startTime := time.Now()

	wg := sync.WaitGroup{}

	for _, phase := range c.Config.Workload.Phases {
		wg.Add(1)
		go func(phase TestPhase) {
			defer wg.Done()

			// wait for phase start time
			<-time.After(phase.StartTime)
			c.l.Info("Starting phase", "Phase", phase.Name, "Type", phase.Type, "Start RPS", phase.StartRPS, "End RPS", phase.EndRPS, "Step", phase.Step, "Duration", phase.Duration)

			switch phase.Type {
			case "constant":
				executor := NewConstantExecutor(client, c.collector, c.funcMgr, c.l)
				executor.Execute(ctx, phase)
			case "variable":
				executor := NewRampingExecutor(client, c.collector, c.funcMgr, c.l)
				executor.Execute(ctx, phase)
			}

		}(phase)
	}

	wg.Wait()

	log.Println("Workload completed in", time.Since(startTime))
	c.collector.Close()
}

func (c *Controller) CreateFunctions() {
	functions := make(map[string]string)

	for _, phase := range c.Config.Workload.Phases {
		functions[phase.ImageTag] = phase.ImageTag
	}

	for _, imageTag := range functions {
		f := c.funcMgr.CreateFunction(imageTag, c.Config.Workload.Timeout, nil,
			&common.Config{
				Memory: 1024 * 1024 * 256, // 256MB
				Cpu: &common.CPUConfig{
					Period: 100000,
					Quota:  100000,
				},
			},
		)
		for i, phase := range c.Config.Workload.Phases {
			if phase.ImageTag == imageTag {
				phase.FunctionID = f.ID
				c.Config.Workload.Phases[i] = phase
			}
		}
	}
}

type Option func(*Controller)

func NewController(logger *slog.Logger, opts ...Option) *Controller {
	c := &Controller{
		l: logger,
	}
	for _, opt := range opts {
		opt(c)
	}

	if c.Config.GenerateWorkload {
		generator := NewWorkloadGenerator(c.Config.Seed, c.Config.MaxDuration, c.Config.LeafAddress, c.Config.Timeout, c.Config.Patterns)
		c.Config.Workload = generator.GenerateWorkload()
	}
	c.funcMgr = NewFunctionManager(c.Config.LeafAddress)
	c.collector = NewCollector()

	return c
}

func WithConfigFile(path string) Option {
	return func(c *Controller) {
		c.Config = &Config{}
		f, err := os.ReadFile(path)
		if err != nil {
			log.Fatal("Failed to read config file:", err)
		}
		err = yaml.Unmarshal(f, c.Config)
		if err != nil {
			log.Fatal("Failed to parse config file:", err)
		}

		if c.Config.LeafAddress == "" {
			log.Fatal("Leaf address is required")
		}

		if c.Config.GenerateWorkload && c.Config.Patterns == nil {
			log.Fatal("Generate workload is true, but no patterns are provided")
		}

		if c.Config.MaxDuration == 0 {
			log.Fatal("Max duration is required")
		}

		if c.Config.Timeout == 0 {
			log.Fatal("Timeout is required")
		}

		if c.Config.Workload != nil {
			for _, phase := range c.Config.Workload.Phases {
				if phase.Type != "constant" && phase.Type != "variable" {
					log.Fatal("Phase type must be either constant or variable")
				}
				if phase.Type == "variable" && (phase.EndRPS == 0 || phase.Step == 0) {
					log.Fatal("Step and end RPS are required for variable phases")
				}
				if phase.Type == "constant" && (phase.StartRPS == 0 || phase.EndRPS != 0 || phase.Step != 0) {
					log.Fatal("Start RPS is required for constant phases")
				}
			}
		}
	}
}
