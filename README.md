# Problem
**How do you impose structure on the input and output of LLM services?**

In Python, there are at least five different ways including outlines, instructor, marginalia, etc. In Go, there are no existing libraries to handle this, so this project presents a proof of concept (POC) using `protojson`.

# Solution
This POC in Go uses protojson for structuring input and output of LLM services.

# Requirements
- `go`
- `protoc` (Protocol Buffers Compiler)
- OpenAI API key set in `OPENAI_API_KEY` environment variable
