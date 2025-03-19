package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"strings"

	"github.com/abiosoft/ishell/v2"
)

var (
	serverAddr = flag.String("p", "", "server address (host:port)")
	helpFlag   = flag.Bool("help", false, "show help")
)

// Store server address (used across multiple functions)
var currentServerAddr string

type NodesListResponse struct {
	Nodes []string `json:"nodes"`
	Error string   `json:"error"`
}

type ConnectResponse struct {
	Status  string            `json:"status"`
	Objects map[string]string `json:"objects"`
	Error   string            `json:"error"`
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

// Connect to server
func connectToServer(address string) (*ConnectResponse, error) {
	url := fmt.Sprintf("http://%s/connect", address)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error connecting to server: %v", err)
	}
	defer resp.Body.Close()

	var result ConnectResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	if result.Error != "" {
		return nil, fmt.Errorf("server error: %s", result.Error)
	}

	return &result, nil
}

// Get list of nodes
func getNodesList(nodeType string) ([]string, error) {
	url := fmt.Sprintf("http://%s/dump/%s", currentServerAddr, nodeType)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error connecting to server: %v", err)
	}
	defer resp.Body.Close()

	var result NodesListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	if result.Error != "" {
		return nil, fmt.Errorf("server error: %s", result.Error)
	}

	return result.Nodes, nil
}

// Get commands for a node type
func getCommandsList(nodeType string) ([]Command, error) {
	url := fmt.Sprintf("http://%s/commands/%s", currentServerAddr, nodeType)
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

func sendCommand(command string, nodeType string, nodeName string) string {
	url := fmt.Sprintf("http://%s/command", currentServerAddr)
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

// Check if a node exists
func checkNodeExists(nodeType string, nodeName string) (bool, error) {
	url := fmt.Sprintf("http://%s/check/%s/%s", currentServerAddr, nodeType, nodeName)
	resp, err := http.Get(url)
	if err != nil {
		return false, fmt.Errorf("error connecting to server: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Exists bool   `json:"exists"`
		Error  string `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("error decoding response: %v", err)
	}

	if result.Error != "" {
		return false, fmt.Errorf("server error: %s", result.Error)
	}

	return result.Exists, nil
}

func setupNodeCommands(shell *ishell.Shell, nodeType string, nodeName string) error {
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
				// If no arguments, show default usage
				if len(c.Args) == 0 {
					c.Println(cmdCopy.DefaultUsage)
					return
				}

				// Build the complete command to send to the server
				command := cmdCopy.Name
				if len(c.Args) > 0 {
					command += " " + strings.Join(c.Args, " ")
				}

				// Send response to server
				response := sendCommand(command, nodeType, nodeName)
				c.Println(response)
			},
		})
	}
	// Add exit command to return to main shell
	shell.AddCmd(&ishell.Cmd{
		Name: "exit",
		Help: "Return to main shell",
		Func: func(c *ishell.Context) {
			c.Stop()
		},
	})

	return nil
}

// Process "select" command
func handleSelectCommand(mainShell *ishell.Shell, nodeType string, nodeName string) {
	// Check if node exists
	exists, err := checkNodeExists(nodeType, nodeName)
	if err != nil {
		mainShell.Println("Error:", err)
		return
	}

	if !exists {
		mainShell.Printf("Node %s of type %s does not exist\n", nodeName, nodeType)
		return
	}

	// Create a sub-shell for this node
	nodeShell := ishell.New()
	nodeShell.SetPrompt(fmt.Sprintf("%s >>> ", nodeName))
	nodeShell.ShowPrompt(true)

	// Setup node commands
	if err := setupNodeCommands(nodeShell, nodeType, nodeName); err != nil {
		mainShell.Println("Error setting up commands:", err)
		return
	}

	// Run the node shell (this will block until exit)
	nodeShell.Run()
}

func showHelp() {
	fmt.Println("Usage:")
	fmt.Println("  client -p <host:port>   : Connect to server")
	fmt.Println("  client -help            : Show this help message")
}

func main() {
	flag.Parse()

	if *helpFlag {
		showHelp()
		return
	}

	if *serverAddr == "" {
		fmt.Println("Server address is required. Use -p <host:port>")
		return
	}

	// Save the server address for use across functions
	currentServerAddr = *serverAddr

	// Connect to server
	connectResp, err := connectToServer(*serverAddr)
	if err != nil {
		fmt.Printf("Failed to connect: %v\n", err)
		return
	}

	// Extract available object types
	var objTypes []string
	for k := range connectResp.Objects {
		objTypes = append(objTypes, k)
	}

	// Create main shell
	shell := ishell.New()
	shell.SetPrompt(">>> ")
	shell.ShowPrompt(true)

	fmt.Printf("Connected to server at %s\n", *serverAddr)
	fmt.Printf("Available object types: %s\n", strings.Join(objTypes, ", "))

	// Add dump command
	shell.AddCmd(&ishell.Cmd{
		Name: "dump",
		Help: "List UEs or gNBs (usage: dump <ue|gnb>)",
		Func: func(c *ishell.Context) {
			if len(c.Args) != 1 {
				c.Println("Usage: dump <ue|gnb>")
				return
			}

			nodeType := c.Args[0]
			nodes, err := getNodesList(nodeType)
			if err != nil {
				c.Println("Error:", err)
				return
			}

			for _, node := range nodes {
				c.Println(node)
			}
		},
	})

	// Add select command
	shell.AddCmd(&ishell.Cmd{
		Name: "select",
		Help: "Select a node to interact with (usage: select <node-name>)",
		Func: func(c *ishell.Context) {
			if len(c.Args) != 1 {
				c.Println("Usage: select <node-name>")
				return
			}

			nodeName := c.Args[0]

			// Try to get the node type from the server
			isUE, err := checkNodeExists("ue", nodeName)
			if err != nil {
				c.Println("Error checking node: %v\n", err)
				return
			}

			if isUE {
				handleSelectCommand(shell, "ue", nodeName)
				return
			}

			isGNB, err := checkNodeExists("gnb", nodeName)
			if err != nil {
				c.Println("Error checking node: %v\n", err)
				return
			}

			if isGNB {
				handleSelectCommand(shell, "gnb", nodeName)
				return
			}
			c.Printf("Node %s does not exist\n", nodeName)
		},
	})

	// Add exit command
	shell.AddCmd(&ishell.Cmd{
		Name: "exit",
		Help: "Exit the client",
		Func: func(c *ishell.Context) {
			c.Stop()
		},
	})

	// Run the shell
	shell.Run()
}
