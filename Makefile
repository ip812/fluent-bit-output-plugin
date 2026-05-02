build:
	@go build -buildmode=c-shared -o ./bin/output_plugin.so .

run:
	@fluent-bit -c fluent-bit.yaml -e ./bin/output_plugin.so
