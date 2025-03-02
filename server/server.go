package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/abiosoft/ishell/v2"
	"github.com/gin-gonic/gin"
)

type Server struct {
	activeUEs []string
	gNodeBs   []string
}

func NewServer() *Server {
	return &Server{
		activeUEs: []string{
			"imsi-306956963543741",
			"imsi-306950959944062",
			"imsi-208937563328413",
			"imsi-208931340068521",
		},
		gNodeBs: []string{
			"MSSIM-gnb-001-01-1",
			"MSSIM-gnb-002-01-1",
			"MSSIM-gnb-003-02-1",
			"MSSIM-gnb-003-03-2",
		},
	}
}

// Setup UE commands
func (s *Server) setupUECommands(shell *ishell.Shell, ueName string) {
	shell.AddCmd(&ishell.Cmd{
		Name: "register",
		Help: "Sign in the UE to Core",
		Func: func(c *ishell.Context) {
			if len(c.Args) == 0 {
				c.Println("Usage: register [--amf] [--smf] [--version] [--help]")
				return
			}
			command := c.Args[0]
			switch command {
			case "--amf":
				c.Println(fmt.Sprintf("Registering UE %s to AMF", ueName))
			case "--smf":
				c.Println(fmt.Sprintf("Registering UE %s to SMF", ueName))
			case "--version":
				c.Println("Register command v1.0")
			case "--help":
				c.Println("Usage: register [--amf] [--smf] [--version] [--help]")
				c.Println("--amf      : Register UE to AMF")
				c.Println("--smf      : Register UE to SMF")
				c.Println("--version  : Show version")
				c.Println("--help     : Show this help message")
			default:
				c.Println("Invalid subcommand for register")
			}
		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "deregister",
		Help: "Logout the UE from Core",
		Func: func(c *ishell.Context) {
			if len(c.Args) == 0 {
				c.Println("Usage: deregister [--force] [--help]")
				return
			}
			command := c.Args[0]
			switch command {
			case "--force":
				c.Println(fmt.Sprintf("Force deregistering UE %s", ueName))
			case "--help":
				c.Println("Usage: deregister [--force]")
				c.Println("--force    : Force deregister UE")
			default:
				c.Println("Invalid subcommand for deregister")
			}
		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "xn-handover",
		Help: "Execute XN handover procedure",
		Func: func(c *ishell.Context) {
			if len(c.Args) == 0 {
				c.Println("Usage: xn-handover --source <gnb-id> --target <gnb-id> [--help] [--version]")
				return
			}
			command := c.Args[0]
			switch command {
			case "--source":
				c.Println(fmt.Sprintf("Executing XN handover for UE %s with source gNB ID %s", ueName, c.Args[1]))
			case "--target":
				c.Println(fmt.Sprintf("Executing XN handover for UE %s with target gNB ID %s", ueName, c.Args[1]))
			case "--help":
				c.Println("Usage: xn-handover --source <gnb-id> --target <gnb-id>")
				c.Println("--source   : Source gNB ID")
				c.Println("--target   : Target gNB ID")
			case "--version":
				c.Println("XN Handover command v1.0")
			default:
				c.Println("Invalid subcommand for xn-handover")
			}
		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "n2-handover",
		Help: "Execute N2 handover procedure",
		Func: func(c *ishell.Context) {
			if len(c.Args) == 0 {
				c.Println("Usage: n2-handover --source <gnb-id> --target <gnb-id> [--help] [--version]")
				return
			}
			command := c.Args[0]
			switch command {
			case "--source":
				c.Println(fmt.Sprintf("Executing N2 handover for UE %s with source gNB ID %s", ueName, c.Args[1]))
			case "--target":
				c.Println(fmt.Sprintf("Executing N2 handover for UE %s with target gNB ID %s", ueName, c.Args[1]))
			case "--help":
				c.Println("Usage: n2-handover --source <gnb-id> --target <gnb-id>")
				c.Println("--source   : Source gNB ID")
				c.Println("--target   : Target gNB ID")
			case "--version":
				c.Println("N2 Handover command v1.0")
			default:
				c.Println("Invalid subcommand for n2-handover")
			}
		},
	})
}

// Setup gNodeB commands
func (s *Server) setupGnbCommands(shell *ishell.Shell, gnbName string) {
	shell.AddCmd(&ishell.Cmd{
		Name: "amf-info",
		Help: "Show some status information about the given AMF",
		Func: func(c *ishell.Context) {
			if len(c.Args) == 0 {
				c.Println("Usage: amf-info [--version] [--help]")
				return
			}
			command := c.Args[0]
			switch command {
			case "--version":
				c.Println("AMF Info Tool v1.0.0")
			case "--help":
				c.Println("Usage: amf-info [--version] [amf-name]")
				c.Println("--version : Show AMF Info version")
				c.Println("--help    : Show this help message")
			default:
				c.Println(fmt.Sprintf("AMF Info for gNodeB %s: Connected and operational", gnbName))
			}
		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "amf-list",
		Help: "List all AMFs associated with the gNB",
		Func: func(c *ishell.Context) {
			if len(c.Args) == 0 || c.Args[0] == "--help" {
				c.Println("Usage: amf-list [--help]")
				c.Println("--help   : Show this help message")
				return
			}
			// Giả sử danh sách AMF là một array mẫu
			amfList := []string{"AMF-01", "AMF-02", "AMF-03"}
			c.Println("AMF List for gNodeB " + gnbName + ":")
			for _, amf := range amfList {
				c.Println(amf)
			}
		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "status",
		Help: "Check the status of the gNodeB",
		Func: func(c *ishell.Context) {
			if len(c.Args) == 0 {
				c.Println("Usage: status [--simple] [--detailed] [--help]")
				return
			}
			command := c.Args[0]
			switch command {
			case "--simple":
				c.Println(fmt.Sprintf("Simple Status for gNodeB %s: Operational", gnbName))
			case "--detailed":
				c.Println(fmt.Sprintf("Detailed Status for gNodeB %s: Full operational details...", gnbName))
			case "--help":
				c.Println("Usage: status [--simple] [--detailed] [--help]")
				c.Println("--simple  : Show simple status")
				c.Println("--detailed : Show detailed status")
			default:
				c.Println("Invalid subcommand for status")
			}
		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "info",
		Help: "Show information about the gNodeB",
		Func: func(c *ishell.Context) {
			if len(c.Args) > 0 && c.Args[0] == "--help" {
				c.Println("Usage: info")
				c.Println("Show information about gNodeB")
				return
			}
			c.Println(fmt.Sprintf("gNodeB Info for %s: Model MSSIM-gnb-001, Location: Data Center A", gnbName))
		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "ue-list",
		Help: "List all UEs associated with the gNodeB",
		Func: func(c *ishell.Context) {
			if len(c.Args) > 0 && c.Args[0] == "--version" {
				c.Println("UE List Tool v1.0.0")
				return
			}
			if len(c.Args) > 0 && c.Args[0] == "--help" {
				c.Println("Usage: ue-list [--version]")
				c.Println("List all UEs associated with the gNodeB")
				return
			}
			// Giả sử danh sách UE kết nối
			ueList := []string{"imsi-306956963543741", "imsi-306950959944062"}
			c.Println("Connected UEs for gNodeB " + gnbName + ":")
			for _, ue := range ueList {
				c.Println(ue)
			}
		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "ue-count",
		Help: "Print the total number of UEs connected to this gNodeB",
		Func: func(c *ishell.Context) {
			if len(c.Args) > 0 && c.Args[0] == "--help" {
				c.Println("Usage: ue-count")
				c.Println("Print the total number of UEs connected to this gNodeB")
				return
			}
			// Giả sử số lượng UE là 2
			ueCount := 2
			c.Println(fmt.Sprintf("Total number of UEs connected to gNodeB %s: %d", gnbName, ueCount))
		},
	})
}

// API Handlers
func (s *Server) handleConnect(c *gin.Context) {
	c.String(200, "Connected successfully")
}

func (s *Server) handleDump(c *gin.Context) {
	response := "Here all the UE and gNodeB server have got:\nUE:\n"
	response += strings.Join(s.activeUEs, "\n")
	response += "\ngNodeB:\n"
	response += strings.Join(s.gNodeBs, "\n")
	c.String(200, response)
}

// Updated executeUECommand function
func (s *Server) executeUECommand(command string, ueName string) string {
	shell := ishell.New()
	s.setupUECommands(shell, ueName)

	// Parse the command to get the command and arguments
	cmdParts := strings.Fields(command)
	if len(cmdParts) == 0 {
		return "Empty command"
	}

	// Create a capture buffer to get the command output
	var outputBuffer strings.Builder
	shell.SetOut(&outputBuffer)

	// Use the correct API to find and execute the command
	err := shell.Process(cmdParts...)
	if err != nil {
		return fmt.Sprintf("Error executing command: %s", err)
	}

	return outputBuffer.String()
}

// Updated executeGnbCommand function
func (s *Server) executeGnbCommand(command string, gnbName string) string {
	shell := ishell.New()
	s.setupGnbCommands(shell, gnbName)

	// Parse the command to get the command and arguments
	cmdParts := strings.Fields(command)
	if len(cmdParts) == 0 {
		return "Empty command"
	}

	// Create a capture buffer to get the command output
	var outputBuffer strings.Builder
	shell.SetOut(&outputBuffer)

	// Use the correct API to find and execute the command
	err := shell.Process(cmdParts...)
	if err != nil {
		return fmt.Sprintf("Error executing command: %s", err)
	}

	return outputBuffer.String()
}

func (s *Server) handleCommand(c *gin.Context) {
	var request struct {
		Command  string `json:"command" binding:"required"`
		NodeType string `json:"nodeType" binding:"required"`
		NodeName string `json:"nodeName" binding:"required"`
	}

	// Bind the request body to the struct
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Handle command based on node type (gnb or ue)
	var response string
	if request.NodeType == "gnb" {
		response = s.executeGnbCommand(request.Command, request.NodeName)
	} else if request.NodeType == "ue" {
		response = s.executeUECommand(request.Command, request.NodeName)
	} else {
		c.JSON(400, gin.H{"error": "Invalid node type"})
		return
	}

	// Send the response
	c.JSON(200, gin.H{"response": response})
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	server := NewServer()

	// Setup routes
	router.GET("/connect", server.handleConnect)
	router.GET("/dump", server.handleDump)
	router.POST("/command", server.handleCommand)

	// Start the HTTP server
	log.Printf("Server starting on port 4000...")
	log.Fatal(router.Run(":4000"))
}
