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
	"sync"

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

// Update response handling for Gin
type CommandResponse struct {
	Response string `json:"response"`
	Error    string `json:"error"`
}

// Fetch active UEs from the server
func getActiveUEs() ([]string, error) {
	url := getServerAddress() + "/active-ues"
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching UEs: %v", err)
	}
	defer resp.Body.Close()

	var result map[string][]string
	body, _ := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling response: %v", err)
	}

	return result["activeUEs"], nil
}

// Fetch gNodeBs from the server
func getGNodeBs() ([]string, error) {
	url := getServerAddress() + "/gnodebs"
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching gNodeBs: %v", err)
	}
	defer resp.Body.Close()

	var result map[string][]string
	body, _ := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling response: %v", err)
	}

	return result["gNodeBs"], nil
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

// Function to send command to multiple UEs concurrently
func sendCommandToMultipleUEs(command string, nodeType string, nodeNames []string) string {
	var responses []string
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, nodeName := range nodeNames {
		wg.Add(1)
		go func(nodeName string) {
			defer wg.Done()
			response := sendCommand(command, nodeType, nodeName)
			mu.Lock()
			responses = append(responses, fmt.Sprintf("Node %s: %s", nodeName, response))
			mu.Unlock()
		}(nodeName)
	}

	wg.Wait()
	return strings.Join(responses, "\n")
}

// Function to send command to multiple gNodeBs concurrently
func sendCommandToMultipleGnb(command string, nodeType string, nodeNames []string) string {
	var responses []string
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, nodeName := range nodeNames {
		wg.Add(1)
		go func(nodeName string) {
			defer wg.Done()
			response := sendCommand(command, nodeType, nodeName)
			mu.Lock()
			responses = append(responses, fmt.Sprintf("Node %s: %s", nodeName, response))
			mu.Unlock()
		}(nodeName)
	}

	wg.Wait()
	return strings.Join(responses, "\n")
}

func setupUECommands(shell *ishell.Shell, ueNames []string) {
	// Add UE commands
	shell.AddCmd(&ishell.Cmd{
		Name: "register",
		Help: "Sign in the UE to Core",
		Func: func(c *ishell.Context) {
			args := strings.Join(c.Args, " ")
			response := sendCommandToMultipleUEs("register "+args, "ue", ueNames)
			c.Println(response)
		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "deregister",
		Help: "Logout the UE from Core",
		Func: func(c *ishell.Context) {
			args := strings.Join(c.Args, " ")
			response := sendCommandToMultipleUEs("deregister "+args, "ue", ueNames)
			c.Println(response)
		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "xn-handover",
		Help: "Execute XN handover procedure",
		Func: func(c *ishell.Context) {
			args := strings.Join(c.Args, " ")
			response := sendCommandToMultipleUEs("xn-handover "+args, "ue", ueNames)
			c.Println(response)
		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "n2-handover",
		Help: "Execute N2 handover procedure",
		Func: func(c *ishell.Context) {
			args := strings.Join(c.Args, " ")
			response := sendCommandToMultipleUEs("n2-handover "+args, "ue", ueNames)
			c.Println(response)
		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "history",
		Help: "Show command history",
		Func: func(c *ishell.Context) {
			c.Println("Command history not implemented yet.")
		},
	})
}

func setupGnbShell(shell *ishell.Shell, gnbNames []string) {
	shell.AddCmd(&ishell.Cmd{
		Name: "amf-info",
		Help: "Show some status information about the given AMF",
		Func: func(c *ishell.Context) {
			args := strings.Join(c.Args, " ")
			response := sendCommandToMultipleGnb("amf-info "+args, "gnb", gnbNames)
			c.Println(response)
		},
	})

	gnbCommands := []struct {
		name string
		help string
	}{
		{"amf-list", "List all AMFs associated with the gNB"},
		{"info", "Show some information about the gNB"},
		{"status", "Show some status information about the gNB"},
		{"ue-count", "Print the total number of UEs connected to this gNB"},
		{"ue-list", "List all UEs associated with the gNB"},
	}

	for _, cmd := range gnbCommands {
		name := cmd.name
		shell.AddCmd(&ishell.Cmd{
			Name: name,
			Help: cmd.help,
			Func: func(c *ishell.Context) {
				args := strings.Join(c.Args, " ")
				response := sendCommandToMultipleGnb(name+" "+args, "gnb", gnbNames)
				c.Println(response)
			},
		})
	}
}

func showHelp() {
	fmt.Println("Usage:")
	fmt.Println("  ./cli -p <port>           : Connect to server")
	fmt.Println("  ./cli --dump              : List all UEs and gNodeBs")
	fmt.Println("  ./cli -ue <active-node>   : Connect to UE")
	fmt.Println("  ./cli -gnb <gnb-name>     : Connect to gNodeB")
}

func main() {
	flag.Parse()

	if *helpFlag {
		showHelp()
		return
	}

	if *port > 0 {
		// Save the port when connecting
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

	if *ueFlag == "" && *gnbFlag == "" {
		fmt.Println("Usage: cli -ue <node-name> or cli -gnb <gnb-name>")
		os.Exit(1)
	}

	// Handling UE interactions
	if *ueFlag != "" {
		var ueNames []string
		if *ueFlag == "" {
			// If no specific UEs provided, get all active UEs
			fetchedUEs, err := getActiveUEs()
			if err != nil {
				fmt.Println("Error fetching active UEs:", err)
				return
			}
			ueNames = fetchedUEs
		} else {
			// Use the specified UEs
			ueNames = strings.Fields(*ueFlag)
		}

		shell := ishell.New()
		shell.SetPrompt(">>> ")
		shell.ShowPrompt(true)

		shell.Println("Connected to UE(s):", strings.Join(ueNames, ", "))
		shell.Println("Type 'help' for available commands")

		setupUECommands(shell, ueNames)
		shell.Run()
		return
	}

	// Handling gNodeB interactions
	if *gnbFlag != "" {
		var gnbNames []string
		if *gnbFlag == "" {
			// If no specific gNodeBs provided, get all gNodeBs
			fetchedGnbs, err := getGNodeBs()
			if err != nil {
				fmt.Println("Error fetching gNodeBs:", err)
				return
			}
			gnbNames = fetchedGnbs
		} else {
			// Use the specified gNodeBs
			gnbNames = strings.Fields(*gnbFlag)
		}

		shell := ishell.New()
		shell.SetPrompt(">>> ")
		shell.ShowPrompt(true)

		shell.Println("Connected to gNodeB(s):", strings.Join(gnbNames, ", "))
		shell.Println("Type 'help' for available commands")

		setupGnbShell(shell, gnbNames)
		shell.Run()
		return
	}

	// If neither UE nor gNodeB is specified, fetch all and interact with all
	if *ueFlag == "" {
		// Fetch the actual list of UEs from the server
		ueNames, err := getActiveUEs()
		if err != nil {
			fmt.Println("Error fetching active UEs:", err)
			return
		}

		shell := ishell.New()
		shell.SetPrompt(">>> ")
		shell.ShowPrompt(true)

		shell.Println("Connected to all UEs:", strings.Join(ueNames, ", "))
		shell.Println("Type 'help' for available commands")

		// Setup commands for UE functionalities
		setupUECommands(shell, ueNames)

		// Start shell
		shell.Run()
	}

	if *gnbFlag == "" {
		// Fetch the actual list of gNodeBs from the server
		gnbNames, err := getGNodeBs()
		if err != nil {
			fmt.Println("Error fetching gNodeBs:", err)
			return
		}

		shell := ishell.New()
		shell.SetPrompt(">>> ")
		shell.ShowPrompt(true)

		shell.Println("Connected to all gNodeBs:", strings.Join(gnbNames, ", "))
		shell.Println("Type 'help' for available commands")

		// Setup commands for gNodeB functionalities
		setupGnbShell(shell, gnbNames)

		// Start shell
		shell.Run()
	}
}
