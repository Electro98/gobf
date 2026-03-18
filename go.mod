module github.com/electro98/gobf

go 1.25.1

replace github.com/electro98/gobf/cli => ./cli

replace github.com/electro98/gobf/interpreter => ./interpreter

require github.com/electro98/gobf/cli v0.0.0-00010101000000-000000000000

require (
	github.com/electro98/gobf/interpreter v0.0.0-00010101000000-000000000000 // indirect
	github.com/urfave/cli/v3 v3.7.0 // indirect
)
