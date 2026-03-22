.PHONY: help setup lint format docs-check

help:
	@echo "DriftLens foundation commands"
	@echo ""
	@echo "make setup      # placeholder setup step"
	@echo "make lint       # placeholder lint step"
	@echo "make format     # placeholder format step"
	@echo "make docs-check # basic docs presence check"

setup:
	@echo "TODO: add toolchain bootstrap (Go, Node, linters) when baseline scaffold begins."

lint:
	@echo "TODO: add real lint commands in baseline scaffold phase."

format:
	@echo "TODO: add real formatting commands in baseline scaffold phase."

docs-check:
	@test -f README.md
	@test -f CONTRIBUTING.md
	@test -f docs/PROJECT-OVERVIEW.md
	@test -f docs/ROADMAP.md
	@test -f docs/ARCHITECTURE-NOTES.md
	@test -f docs/WORKFLOW.md
	@echo "Docs check passed."
