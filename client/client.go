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

// Add this new function to save and load port
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
	fmt.Println("  ./cli -ue <active-node>   : Connect to UE")
	fmt.Println("  ./cli -gnb <gnb-name>     : Connect to gNodeB")
}

func setupGnbShell(shell *ishell.Shell, gnbName string) {
	shell.AddCmd(&ishell.Cmd{
		Name: "amf-info",
		Help: "Show some status information about the given AMF",
		Func: func(c *ishell.Context) {
			args := strings.Join(c.Args, " ")
			response := sendCommand("amf-info "+args, "gnb", gnbName)
			c.Println(response)
		},
	})

	// Add other gNB commands
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
				response := sendCommand(name+" "+args, "gnb", gnbName)
				c.Println(response)
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

	if *ueFlag != "" {
		shell := ishell.New()
		shell.SetPrompt(">>> ")
		shell.ShowPrompt(true)

		shell.Println("Connected to node:", *ueFlag)
		shell.Println("Type 'help' for available commands")

		// Add base commands
		shell.AddCmd(&ishell.Cmd{
			Name: "help",
			Help: "Display available commands",
			Func: func(c *ishell.Context) {
				response := sendCommand("help", "ue", *ueFlag)
				c.Println(response)
			},
		})

		shell.AddCmd(&ishell.Cmd{
			Name: "register",
			Help: "Sign in the UEs to Core",
			Func: func(c *ishell.Context) {
				args := strings.Join(c.Args, " ")
				response := sendCommand("register "+args, "ue", *ueFlag)
				c.Println(response)
			},
		})

		shell.AddCmd(&ishell.Cmd{
			Name: "deregister",
			Help: "Logout the UEs from Core",
			Func: func(c *ishell.Context) {
				args := strings.Join(c.Args, " ")
				response := sendCommand("deregister "+args, "ue", *ueFlag)
				c.Println(response)
			},
		})

		shell.AddCmd(&ishell.Cmd{
			Name: "xn-handover",
			Help: "Execute XN handover procedure",
			Func: func(c *ishell.Context) {
				args := strings.Join(c.Args, " ")
				response := sendCommand("xn-handover "+args, "ue", *ueFlag)
				c.Println(response)
			},
		})

		shell.AddCmd(&ishell.Cmd{
			Name: "n2-handover",
			Help: "Execute N2 handover procedure",
			Func: func(c *ishell.Context) {
				args := strings.Join(c.Args, " ")
				response := sendCommand("n2-handover "+args, "ue", *ueFlag)
				c.Println(response)
			},
		})

		shell.AddCmd(&ishell.Cmd{
			Name: "history",
			Help: "Show command history",
			Func: func(c *ishell.Context) {
				response := getHistory()
				c.Println(response)
			},
		})

		// Start shell
		shell.Run()
		return
	}

	if *gnbFlag != "" {
		shell := ishell.New()
		shell.SetPrompt(">>> ")
		shell.ShowPrompt(true)

		shell.Println("Connected to gNodeB:", *gnbFlag)
		shell.Println("Type 'help' for available commands")

		setupGnbShell(shell, *gnbFlag)
		shell.Run()
		return
	}

	fmt.Println("Invalid command. Use --help to see usage.")
}

func getHistory() any {
	panic("unimplemented")
}
