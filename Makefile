build-plugins:
  go build -buildmode=plugin -o mock.so internal/assistant/plugins/copilot/mock/mock_copilot_plugin.go && \
	go build -buildmode=plugin -o gemini.so internal/assistant/plugins/copilot/gemini/gemini_copilot_plugin.go
gen:
	@command -v buf >/dev/null 2>&1 || { echo "Buf is not installed. Install it with 'brew install buf'."; exit 1; }
	buf generate

clean:
	@echo "Cleaning generated files..."
	rm -rf gen
