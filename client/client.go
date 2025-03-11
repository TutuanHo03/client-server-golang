package main

import (
	"bytes"
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
	port     = flag.Int("p", 0, "server port")
	dump     = flag.Bool("dump", false, "list all UEs and gNodeBs")
	helpFlag = flag.Bool("help", false, "show help")
	ueFlag   = flag.String("ue", "", "connect to UE node")
	gnbFlag  = flag.String("gnb", "", "connect to gNodeB")
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

type Command struct {
	Name         string `json:"name"`
	Help         string `json:"help"`
	DefaultUsage string `json:"defaultUsage"`
}

type CommandsListResponse struct {
	Commands []Command `json:"commands"`
	Error    string    `json:"error"`
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

func getCommandsList(nodeType string) ([]Command, error) {
	url := fmt.Sprintf("%s/commands/%s", getServerAddress(), nodeType)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error connecting to server: %v", err)
	}
	defer resp.Body.Close()

	var result CommandsListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	if result.Error != "" {
		return nil, fmt.Errorf("server error: %s", result.Error)
	}

	return result.Commands, nil
}

func showHelp() {
	fmt.Println("Usage:")
	fmt.Println("  ./cli -p <port>           : Connect to server")
	fmt.Println("  ./cli --dump              : List all UEs and gNodeBs")
	fmt.Println("  ./cli -ue <ue-name>       : Connect to a specific UE")
	fmt.Println("  ./cli -gnb <gnb-name>     : Connect to a specific gNodeB")
}

func setupCommands(shell *ishell.Shell, nodeType string, nodeNames []string) error {
	// Get the list of commands from server
	commands, err := getCommandsList(nodeType)
	if err != nil {
		return fmt.Errorf("failed to get commands: %v", err)
	}

	// Add each command into shell
	for _, cmd := range commands {
		cmdCopy := cmd // create the replica for closure

		shell.AddCmd(&ishell.Cmd{
			Name: cmdCopy.Name,
			Help: cmdCopy.Help,
			Func: func(c *ishell.Context) {
				// Nếu không có tham số, hiển thị cách sử dụng
				if len(c.Args) == 0 {
					c.Println(cmdCopy.DefaultUsage)
					return
				}

				// Build the complete command to send to the server
				for _, nodeName := range nodeNames {
					command := cmdCopy.Name
					if len(c.Args) > 0 {
						command += " " + strings.Join(c.Args, " ")
					}

					// Send response to server
					response := sendCommand(command, nodeType, nodeName)

					// Display the response
					if len(nodeNames) > 1 {
						if nodeType == "ue" {
							c.Printf("Response for UE %s:\n%s\n", nodeName, response)
						} else {
							c.Printf("Response for gNodeB %s:\n%s\n", nodeName, response)
						}
					} else {
						c.Println(response)
					}
				}
			},
		})
	}

	return nil
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

	shell := ishell.New()
	shell.SetPrompt(">>> ")
	shell.ShowPrompt(true)

	if *ueFlag != "" {
		ueNames := strings.Fields(*ueFlag)
		shell.Println("Connected to UE(s):", strings.Join(ueNames, ", "))
		shell.Println("Type 'help' for available commands")

		// Get the commands of UE from server and setup shell
		if err := setupCommands(shell, "ue", ueNames); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		shell.Run()
	} else if *gnbFlag != "" {
		gnbNames := strings.Fields(*gnbFlag)
		shell.Println("Connected to gNodeB(s):", strings.Join(gnbNames, ", "))
		shell.Println("Type 'help' for available commands")

		// Get the commands of Gnb from server and setup shell
		if err := setupCommands(shell, "gnb", gnbNames); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		shell.Run()
	} else {
		fmt.Println("Usage: cli -ue <ue-name> or cli -gnb <gnb-name>")
		os.Exit(1)
	}
}
