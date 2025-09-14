# Define the Go compiler
GO := go

# Define the build flags for plugins
PLUGIN_FLAGS := -buildmode=plugin

# Define the source files and their corresponding output names
PLUGINS := gemini.so mock.so

# Define the source paths for the plugins
PLUGIN_SRCS := \
	internal/assistant/plugins/copilot/gemini/gemini_copilot_plugin.go \
	internal/assistant/plugins/copilot/mock/mock_copilot_plugin.go

# A map of output names to source paths
PLUGIN_MAP := gemini.so=internal/assistant/plugins/copilot/gemini/gemini_copilot_plugin.go \
              mock.so=internal/assistant/plugins/copilot/mock/mock_copilot_plugin.go

#------------------------------------------------------------------------------
# Targets
#------------------------------------------------------------------------------

.PHONY: all build-plugins gen clean

all: build-plugins gen

build-plugins: $(PLUGINS)

# A pattern rule to build each plugin dynamically
%.so:
	@echo "Building plugin: $@"
	$(GO) build $(PLUGIN_FLAGS) -o $@ $(word 1,$(subst $@=,,$(PLUGIN_MAP)))

gen:
	@command -v buf >/dev/null 2>&1 || { echo "Buf is not installed. Install it with 'brew install buf'."; exit 1; }
	buf generate

clean:
	@echo "Cleaning generated files..."
	rm -rf gen $(PLUGINS)
