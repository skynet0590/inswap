package core

import "github.com/skynet0590/inswap/app"

// ServerCore is the core of server background. It controls all the components of the server
type ServerCore struct {

}

// CoreConf is the configuration data required to create a new ServerCore
type CoreConf struct {
	DataDir           string
	Network           app.Network
}
