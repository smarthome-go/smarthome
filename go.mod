module github.com/smarthome-go/smarthome

go 1.21

toolchain go1.22.0

require (
	github.com/briandowns/openweathermap v0.19.0
	github.com/davecgh/go-spew v1.1.1
	github.com/go-co-op/gocron v1.28.3
	github.com/go-ping/ping v1.1.0
	github.com/go-sql-driver/mysql v1.7.1
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/securecookie v1.1.1
	github.com/gorilla/sessions v1.2.1
	github.com/gorilla/websocket v1.5.0
	github.com/h2non/filetype v1.1.3
	github.com/lnquy/cron v1.1.1
	github.com/nathan-osman/go-sunrise v1.1.0
	github.com/rifflock/lfshook v0.0.0-20180920164130-b9218ef580f5
	github.com/sirupsen/logrus v1.9.3
	github.com/smarthome-go/homescript/v3 v3.0.0-00010101000000-000000000000
	github.com/stretchr/testify v1.8.2
	golang.org/x/crypto v0.10.0
	golang.org/x/exp v0.0.0-20230522175609-2e198f4a06a1
)

require (
	github.com/agnivade/levenshtein v1.1.1 // indirect
	github.com/google/uuid v1.2.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	golang.org/x/net v0.10.0 // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
	golang.org/x/sys v0.9.0 // indirect
	golang.org/x/text v0.10.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/smarthome-go/homescript/v3 => ../homescript/
