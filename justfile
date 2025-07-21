test-reducing:
    go run cmd/main.go --config=test/configs/reducing_config.yaml

test-overlapping:
    go run cmd/main.go --config=test/configs/overlapping.yaml

test-generate-big:
    go run cmd/main.go --config=test/configs/generate-big-config.yaml

test-generate-small:
    go run cmd/main.go --config=test/configs/generate-small-config.yaml

howmanyerrors:
    awk -F',' '{if($5!="") count++} END{print count}' results.csv

run config_file:
    go run cmd/main.go --config={{config_file}} --log-level=info


###### pprofing ######
pprof:
    go tool pprof -http=:9092 http://localhost:6060/debug/pprof/profile

heap:
    go tool pprof -http=:9093 http://localhost:6060/debug/pprof/heap

goroutine:
    go tool pprof -http=:9094 http://localhost:6060/debug/pprof/goroutine