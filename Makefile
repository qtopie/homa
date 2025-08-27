gen:
	@command -v buf >/dev/null 2>&1 || { echo "Buf is not installed. Install it with 'brew install buf'."; exit 1; }
	buf generate

clean:
	@echo "Cleaning generated files..."
	rm -rf gen
