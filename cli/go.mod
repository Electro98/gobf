module github.com/electro98/gobf/cli

go 1.25.1

replace github.com/electro98/gobf/interpreter => ../interpreter

require (
	github.com/electro98/gobf/interpreter v0.0.0-00010101000000-000000000000
	github.com/urfave/cli/v3 v3.7.0
)
