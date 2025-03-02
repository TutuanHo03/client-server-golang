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

// Response structure for command results
type CommandResponse struct {
	Response string `json:"response"`
	Error    string `json:"error"`
}

// Send a single command to a specific node
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
}

// Setup UE structure commands
func setupUEShell(shell *ishell.Shell, ueNames []string) {
	for _, ueName := range ueNames {
		shell.AddCmd(&ishell.Cmd{
			Name: "register",
			Help: "Sign in the UE to Core",
			Func: func(c *ishell.Context) {
				args := strings.Join(c.Args, " ")
				response := sendCommand("register "+args, "ue", ueName)
				c.Println(response)
			},
		})

		shell.AddCmd(&ishell.Cmd{
			Name: "deregister",
			Help: "Logout the UE from Core",
			Func: func(c *ishell.Context) {
				args := strings.Join(c.Args, " ")
				response := sendCommand("deregister "+args, "ue", ueName)
				c.Println(response)
			},
		})

		shell.AddCmd(&ishell.Cmd{
			Name: "xn-handover",
			Help: "Execute XN handover procedure",
			Func: func(c *ishell.Context) {
				args := strings.Join(c.Args, " ")
				response := sendCommand("xn-handover "+args, "ue", ueName)
				c.Println(response)
			},
		})

		shell.AddCmd(&ishell.Cmd{
			Name: "n2-handover",
			Help: "Execute N2 handover procedure",
			Func: func(c *ishell.Context) {
				args := strings.Join(c.Args, " ")
				response := sendCommand("n2-handover "+args, "ue", ueName)
				c.Println(response)
			},
		})
	}
}

// Setup gNodeB structure commands
func setupGnbShell(shell *ishell.Shell, gnbNames []string) {
	for _, gnbName := range gnbNames {
		shell.AddCmd(&ishell.Cmd{
			Name: "amf-info",
			Help: "Show status information about AMF",
			Func: func(c *ishell.Context) {
				args := strings.Join(c.Args, " ")
				response := sendCommand("amf-info "+args, "gnb", gnbName)
				c.Println(response)
			},
		})

		shell.AddCmd(&ishell.Cmd{
			Name: "amf-list",
			Help: "List all AMFs associated with the gNB",
			Func: func(c *ishell.Context) {
				args := strings.Join(c.Args, " ")
				response := sendCommand("amf-list "+args, "gnb", gnbName)
				c.Println(response)
			},
		})

		shell.AddCmd(&ishell.Cmd{
			Name: "status",
			Help: "Check the status of the gNodeB",
			Func: func(c *ishell.Context) {
				args := strings.Join(c.Args, " ")
				response := sendCommand("status "+args, "gnb", gnbName)
				c.Println(response)
			},
		})

		shell.AddCmd(&ishell.Cmd{
			Name: "info",
			Help: "Show information about the gNodeB",
			Func: func(c *ishell.Context) {
				args := strings.Join(c.Args, " ")
				response := sendCommand("info "+args, "gnb", gnbName)
				c.Println(response)
			},
		})

		shell.AddCmd(&ishell.Cmd{
			Name: "ue-list",
			Help: "List all UEs associated with the gNodeB",
			Func: func(c *ishell.Context) {
				args := strings.Join(c.Args, " ")
				response := sendCommand("ue-list "+args, "gnb", gnbName)
				c.Println(response)
			},
		})

		shell.AddCmd(&ishell.Cmd{
			Name: "ue-count",
			Help: "Print the total number of UEs connected to this gNodeB",
			Func: func(c *ishell.Context) {
				args := strings.Join(c.Args, " ")
				response := sendCommand("ue-count "+args, "gnb", gnbName)
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

	// If dump flag is provided, retrieve the dump from the server
	if *dump {
		response := getDump()
		fmt.Println(response)
		return
	}

	// If user specifies ueFlag, set up UE commands
	if *ueFlag != "" {
		ueNames := strings.Fields(*ueFlag)

		shell := ishell.New()
		shell.SetPrompt(">>> ")
		shell.ShowPrompt(true)

		shell.Println("Connected to UE(s):", strings.Join(ueNames, ", "))
		shell.Println("Type 'help' for available commands")

		setupUEShell(shell, ueNames)
		shell.Run()
		return
	}

	// If user specifies gnbFlag, set up gNodeB commands
	if *gnbFlag != "" {
		gnbNames := strings.Fields(*gnbFlag)

		shell := ishell.New()
		shell.SetPrompt(">>> ")
		shell.ShowPrompt(true)

		shell.Println("Connected to gNodeB(s):", strings.Join(gnbNames, ", "))
		shell.Println("Type 'help' for available commands")

		setupGnbShell(shell, gnbNames)
		shell.Run()
		return
	}

	// If neither -ue nor -gnb is provided, show usage
	fmt.Println("Usage: cli -ue <ue-name> or cli -gnb <gnb-name>")
	os.Exit(1)
}
