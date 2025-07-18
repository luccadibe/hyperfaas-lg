# HyperFaaS Load Generator

A load testing tool for HyperFaaS serverless functions with support for manual workload definition and automatic workload generation.

## Usage

```bash
go run cmd/main.go --config=test/configs/config.yaml --log-level=info
```

## Configuration

### Manual Workload

Define explicit test phases:

```yaml
leaf_address: localhost:50050
max_duration: 30s
timeout: 10
workload:
  phases:
    - name: phase1
      type: constant        # Fixed RPS
      start_time: 1s
      start_rps: 10
      duration: 10s
      image_tag: hyperfaas-echo:latest
    - name: phase2
      type: variable        # Ramping RPS
      start_time: 15s
      start_rps: 10
      end_rps: 100
      step: 5              # RPS increment
      duration: 10s
      image_tag: hyperfaas-echo:latest
```

### Generated Workload

Define patterns for automatic workload generation:

```yaml
leaf_address: localhost:50050
max_duration: 30s
timeout: 10
generate_workload: true
seed: 123
patterns:
  echo-workloads:
    image_tag: hyperfaas-echo:latest
    phase_count:
      min: 2
      max: 5
    constant_likelihood: 0.6    # 60% chance of constant phases
    ramping_likelihood: 0.4     # 40% chance of variable phases
    parameters:
      start_rps:
        min: 10
        max: 50
      end_rps:
        min: 75
        max: 150
      step:
        min: 5
        max: 20
```

## How It Works

1. **Controller** loads config and creates HyperFaaS functions
2. **WorkloadGenerator** (if enabled) creates random phases from patterns
3. **Executors** run phases in parallel:
   - `ConstantExecutor`: Maintains fixed RPS
   - `RampingExecutor`: Gradually increases/decreases RPS
4. **Collector** gathers performance metrics
5. All phases execute concurrently based on their `start_time`

## Phase Types

- **constant**: Fixed RPS for the entire duration
- **variable**: RPS changes from `start_rps` to `end_rps` in `step` increments

Multiple patterns can run overlapping phases.
