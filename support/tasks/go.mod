module github.com/netspective-media/news.healthcareguys.com/support/tasks

go 1.12

require (
	github.com/lectio/link v0.0.0-20190519022856-f1562762c35a
	github.com/lectio/markdown v0.0.0-20190521141033-e469175595b8
	github.com/lectio/progress v0.0.0-20190522232809-f0a219a98ea5
	github.com/lectio/publish v0.0.0-20190522234741-80587a488c1e
	github.com/lectio/resource v0.0.0-20190524200617-96910fee2dab // indirect
	github.com/mattn/go-colorable v0.1.2 // indirect
	golang.org/x/sys v0.0.0-20190524152521-dbbf3f1254d4 // indirect
	golang.org/x/tools v0.0.0-20190524184802-d487f80763e2 // indirect
)

replace github.com/lectio/publish v0.0.0-20190522234741-80587a488c1e => ../publish
