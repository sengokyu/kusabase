module github.com/sengokyu/kusabase/cli

go 1.25.0

require (
	github.com/sengokyu/kusabase/httpclient v0.0.0-00010101000000-000000000000
	github.com/spf13/cobra v1.10.2
	golang.org/x/term v0.40.0
)

require (
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.9 // indirect
	golang.org/x/sys v0.41.0 // indirect
)

replace github.com/sengokyu/kusabase/httpclient => ../httpclient
