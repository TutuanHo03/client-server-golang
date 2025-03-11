package main

import (
	"bytes"
	"client-server/models"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/abiosoft/ishell/v2"
)

var (
	port        = flag.Int("p", 0, "server port")
	dump        = flag.Bool("dump", false, "list all UEs and gNodeBs")
	helpFlag    = flag.Bool("help", false, "show help")
	ueFlag      = flag.String("ue", "", "connect to UE node")
	gnbFlag     = flag.String("gnb", "", "connect to gNodeB")
	commandFile = flag.String("c", "", "path to command configuration JSON file")
)

func savePort(port int) error {
	return ioutil.WriteFile(".port", []byte(fmt.Sprintf("%d", port)), 0644)
}

func loadPort() int {
	data, err := ioutil.ReadFile(".port")
	if err != nil {
		return 0
	}
	port, _ := strconv.Atoi(string(data))
	return port
}

func getServerAddress() string {
	savedPort := loadPort()
	if *port > 0 {
		savedPort = *port
	}
	return fmt.Sprintf("http://localhost:%d", savedPort)
}

func getDump() string {
	url := getServerAddress() + "/dump"
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	return string(body)
}

type CommandResponse struct {
	Response string `json:"response"`
	Error    string `json:"error"`
}

func sendCommand(command string, nodeType string, nodeName string) string {
	url := getServerAddress() + "/command"
	data, _ := json.Marshal(map[string]string{
		"command":  command,
		"nodeType": nodeType,
		"nodeName": nodeName,
	})

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}
	defer resp.Body.Close()

	var result CommandResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Sprintf("Error decoding response: %v", err)
	}

	if result.Error != "" {
		return result.Error
	}
	return result.Response
}

func showHelp() {
	fmt.Println("Usage:")
	fmt.Println("  ./cli -p <port>           : Connect to server")
	fmt.Println("  ./cli --dump              : List all UEs and gNodeBs")
	fmt.Println("  ./cli -ue <ue-name>       : Connect to a specific UE")
	fmt.Println("  ./cli -gnb <gnb-name>     : Connect to a specific gNodeB")
	fmt.Println("  ./cli -c <config-file>    : Load commands from a JSON configuration file")
}

func loadCommandsConfig(filePath string) (*models.CommandConfig, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %v", err)
	}

	var commands models.CommandConfig
	err = json.Unmarshal(data, &commands)
	if err != nil {
		return nil, fmt.Errorf("error parsing config file: %v", err)
	}

	return &commands, nil
}

func setupUECommands(shell *ishell.Shell, commandsConf *models.CommandConfig, ueNames []string) {
	for _, cmd := range commandsConf.UE.Commands {
		// Create a copy of cmd for the closure
		cmdCopy := cmd

		shell.AddCmd(&ishell.Cmd{
			Name: cmdCopy.Name,
			Help: cmdCopy.Help,
			Func: func(c *ishell.Context) {
				for _, ueName := range ueNames {
					command := cmdCopy.Name
					if len(c.Args) > 0 {
						command += " " + strings.Join(c.Args, " ")
					}
					response := sendCommand(command, "ue", ueName)
					if len(ueNames) > 1 {
						c.Printf("Response for UE %s:\n%s\n", ueName, response)
					} else {
						c.Println(response)
					}
				}
			},
		})
	}
}

func setupGnbCommands(shell *ishell.Shell, commandsConf *models.CommandConfig, gnbNames []string) {
	for _, cmd := range commandsConf.GNB.Commands {
		// Create a copy of cmd for the closure
		cmdCopy := cmd

		shell.AddCmd(&ishell.Cmd{
			Name: cmdCopy.Name,
			Help: cmdCopy.Help,
			Func: func(c *ishell.Context) {
				for _, gnbName := range gnbNames {
					command := cmdCopy.Name
					if len(c.Args) > 0 {
						command += " " + strings.Join(c.Args, " ")
					}
					response := sendCommand(command, "gnb", gnbName)
					if len(gnbNames) > 1 {
						c.Printf("Response for gNodeB %s:\n%s\n", gnbName, response)
					} else {
						c.Println(response)
					}
				}
			},
		})
	}
}

func main() {
	flag.Parse()

	if *helpFlag {
		showHelp()
		return
	}

	if *port > 0 {
		if err := savePort(*port); err != nil {
			fmt.Printf("Error saving port: %v\n", err)
			return
		}
		_, err := http.Get(getServerAddress() + "/connect")
		if err != nil {
			fmt.Printf("Error connecting to server: %v\n", err)
			return
		}
		fmt.Printf("Connected to port %d successfully\n", *port)
		return
	}

	if loadPort() == 0 {
		fmt.Println("Not connected to any port. Please use -p <port> first")
		return
	}

	if *dump {
		response := getDump()
		fmt.Println(response)
		return
	}

	// Check if command file is provided
	if *commandFile == "" {
		fmt.Println("Command configuration file is required. Use -c <config-file>")
		return
	}

	// Load commands from file
	commandsConf, err := loadCommandsConfig(*commandFile)
	if err != nil {
		fmt.Printf("Error loading commands: %v\n", err)
		return
	}

	shell := ishell.New()
	shell.SetPrompt(">>> ")
	shell.ShowPrompt(true)

	if *ueFlag != "" {
		ueNames := strings.Fields(*ueFlag)
		shell.Println("Connected to UE(s):", strings.Join(ueNames, ", "))
		shell.Println("Type 'help' for available commands")
		setupUECommands(shell, commandsConf, ueNames)
		shell.Run()
	} else if *gnbFlag != "" {
		gnbNames := strings.Fields(*gnbFlag)
		shell.Println("Connected to gNodeB(s):", strings.Join(gnbNames, ", "))
		shell.Println("Type 'help' for available commands")
		setupGnbCommands(shell, commandsConf, gnbNames)
		shell.Run()
	} else {
		fmt.Println("Usage: cli -ue <ue-name> or cli -gnb <gnb-name>")
		os.Exit(1)
	}
}
