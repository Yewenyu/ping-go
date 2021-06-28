module github.com/Yewenyu/ping-go

go 1.16

require (
	github.com/cloverstd/tcping v0.1.1
	github.com/go-ping/ping v0.0.0-20210506233800-ff8be3320020
	github.com/smartystreets/goconvey v1.6.4 // indirect
	golang.org/x/net v0.0.0-20210610132358-84b48f89b13b // indirect
	golang.org/x/sys v0.0.0-20210608053332-aa57babbf139 // indirect
)

replace github.com/cloverstd/tcping => ../tcping-master
