package gam

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/fatih/color"
	"github.com/maldan/gam/internal/app/gam/app"
	"github.com/maldan/gam/internal/app/gam/core"
	"github.com/maldan/gam/internal/app/gam/process"
	"github.com/maldan/go-cmhp/cmhp_file"
)

type Command struct {
	Description string
	Execute     func(p ...string)
	Params      int
}

func SetStructField(s interface{}, field string, value interface{}) {
	mappo := make(map[string]interface{})
	bb, _ := json.Marshal(s)
	json.Unmarshal(bb, &mappo)
	mappo[field] = value
	b2, _ := json.Marshal(&mappo)
	json.Unmarshal(b2, s)
}

func Start(version string) {
	// No args
	if len(os.Args) <= 1 {
		core.Exit("No params")
	}

	// Create command list
	commandList := make(map[string]Command)

	// Init command
	commandList["init"] = Command{
		Description: "Init and install gam into system",
		Execute: func(p ...string) {
			// Copy program
			source, _ := os.Open(os.Args[0])
			destination, err := os.Create(core.GamDir + "/gam")
			if err != nil {
				panic(err)
			}
			io.Copy(destination, source)

			// Set executable
			os.Chmod(core.GamDir+"/gam", 0755)

			// Create script
			d1 := []byte("#!/bin/bash\n~/.gam/gam \"$@\"")
			ioutil.WriteFile("/usr/local/bin/gam", d1, 0755)
		},
	}

	// Help command
	commandList["help"] = Command{
		Description: "Show help info",
		Execute: func(p ...string) {
			for k, v := range commandList {
				fmt.Print(color.GreenString(k) + " ")
				for i := 0; i < v.Params; i++ {
					fmt.Print(color.YellowString(fmt.Sprintf("$%v", i)) + " ")
				}

				fmt.Println("\n - " + v.Description)
			}
		},
	}

	// Install app
	commandList["i"] = Command{
		Params:      1,
		Description: "Install app from $0",
		Execute: func(p ...string) {
			app.Install(p[0])
		},
	}

	// Application list
	commandList["al"] = Command{
		Description: "List of installed applications",
		Execute: func(p ...string) {
			app.List()
		},
	}

	// Delete application
	commandList["delete"] = Command{
		Params:      1,
		Description: "Delete application $0",
		Execute: func(p ...string) {
			app.Delete(p[0])
		},
	}

	// Process list
	commandList["pl"] = Command{
		Description: "Process list",
		Execute: func(p ...string) {
			pl := process.GamList()
			for _, p := range pl {
				fmt.Printf("pid: %v\n", p.Pid)
				fmt.Printf("cmd: %v\n", p.Cmd)
				for k, v := range p.Args {
					fmt.Printf("%v: %v\n", k, v)
				}
				fmt.Println()
			}
		},
	}

	// Run app
	commandList["run"] = Command{
		Params:      1,
		Description: "Run app $0",
		Execute: func(p ...string) {
			app.Run(p[0], p[1:])
		},
	}

	// Kill process
	commandList["kill"] = Command{
		Params:      1,
		Description: "Kill process $0",
		Execute: func(p ...string) {
			process.Kill(p[0])
		},
	}

	// Kill process
	commandList["version"] = Command{
		Description: "Version",
		Execute: func(p ...string) {
			fmt.Println(version)
		},
	}

	// Print config
	commandList["pcfg"] = Command{
		Description: "Print config",
		Execute: func(p ...string) {
			fmt.Println("DefaultHost:", core.GamConfig.DefaultHost)
		},
	}

	// Set variable
	commandList["set"] = Command{
		Params:      2,
		Description: "Set variable $0 = $1",
		Execute: func(p ...string) {
			SetStructField(&core.GamConfig, p[0], p[1])
			cmhp_file.WriteJSON(core.GamDataDir+"/gam/config.json", &core.GamConfig)
			commandList["pcfg"].Execute()
		},
	}

	// Check command
	if _, ok := commandList[os.Args[1]]; !ok {
		core.Exit("Unknown command: " + os.Args[1])
	}

	// Check params
	if len(os.Args[2:]) < commandList[os.Args[1]].Params {
		core.Exit("Not enough params for: " + os.Args[1])
	}

	// Run command
	commandList[os.Args[1]].Execute(os.Args[2:]...)
}
