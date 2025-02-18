package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
)

type UECommand struct {
	Description string
	Handler     func(args []string, nodeName string) string
	SubCommands map[string]string
}

type Server struct {
	activeUEs  []string
	gNodeBs    []string
	ueCommands map[string]UECommand
}

func NewServer() *Server {
	s := &Server{
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

	// Initialize UE commands
	s.ueCommands = map[string]UECommand{
		"help": {
			Description: "Display available commands",
			Handler: func(args []string, nodeName string) string {
				var help strings.Builder
				help.WriteString("Available commands:\n")
				for name, cmd := range s.ueCommands {
					help.WriteString(fmt.Sprintf("%-12s| %s\n", name, cmd.Description))
				}
				return help.String()
			},
		},
		"register": {
			Description: "Sign in the UE to Core",
			Handler: func(args []string, nodeName string) string {
				// Let handleUeCommand handle the --help flag
				return fmt.Sprintf("Registering UE %s with args: %s", nodeName, strings.Join(args, " "))
			},
			SubCommands: map[string]string{
				"--amf":     "Register UE to AMF",
				"--smf":     "Register UE to SMF",
				"--help":    "Usage: register [--amf] [--smf]\nRegister UE to core network components",
				"--version": "Show command version",
			},
		},
		"deregister": {
			Description: "Logout the UE from Core",
			Handler: func(args []string, nodeName string) string {
				// Let handleUeCommand handle the --help flag
				return fmt.Sprintf("Deregistering UE %s with args: %s", nodeName, strings.Join(args, " "))
			},
			SubCommands: map[string]string{
				"--force": "Force deregister",
				"--help":  "Usage: deregister [--force]\nDeregister UE from core network",
			},
		},
		"xn-handover": {
			Description: "Execute XN handover procedure",
			Handler: func(args []string, nodeName string) string {
				// Let handleUeCommand handle the --help flag
				return fmt.Sprintf("Executing XN handover for UE %s with args: %s", nodeName, strings.Join(args, " "))
			},
			SubCommands: map[string]string{
				"--source": "Source gNB ID",
				"--target": "Target gNB ID",
				"--help":   "Usage: xn-handover --source <gnb-id> --target <gnb-id>\nExecute XN handover between gNodeBs",
			},
		},
		"n2-handover": {
			Description: "Execute N2 handover procedure",
			Handler: func(args []string, nodeName string) string {
				// Let handleUeCommand handle the --help flag
				return fmt.Sprintf("Executing N2 handover for UE %s with args: %s", nodeName, strings.Join(args, " "))
			},
			SubCommands: map[string]string{
				"--source": "Source gNB ID",
				"--target": "Target gNB ID",
				"--help":   "Usage: n2-handover --source <gnb-id> --target <gnb-id>\nExecute N2 handover between gNodeBs",
			},
		},
	}

	return s
}

// Convert handlers to Gin handlers
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

type CommandRequest struct {
	Command  string `json:"command" binding:"required"`
	NodeType string `json:"nodeType" binding:"required"`
	NodeName string `json:"nodeName" binding:"required"`
}

func (s *Server) handleCommand(c *gin.Context) {
	var request CommandRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var response string
	if request.NodeType == "gnb" {
		response = s.handleGnbCommand(request.Command, request.NodeName)
	} else {
		response = s.handleUeCommand(request.Command, request.NodeName)
	}

	c.JSON(200, gin.H{"response": response})
}

func (s *Server) handleGnbCommand(cmd string, gnbName string) string {
	parts := strings.Fields(cmd)
	command := parts[0]
	args := parts[1:]

	switch command {
	case "help":
		return `Available commands:
			amf-info  | Show some status information about the given AMF
			amf-list  | List all AMFs associated with the gNB
			info      | Show some information about the gNB
			status    | Show some status information about the gNB
			ue-count  | Print the total number of UEs connected the this gNB
			ue-list   | List all UEs associated with the gNB`

	case "amf-info":
		if len(args) > 0 && args[0] == "--help" {
			return "Usage: amf-info [--version] [amf-name]\nShow detailed information about specified AMF"
		}
		if len(args) > 0 && args[0] == "--version" {
			return "AMF Info Tool v1.0.0"
		}
		return fmt.Sprintf("AMF Info for %s: Connected and operational", gnbName)

	case "ue-list":
		if len(args) > 0 && args[0] == "--version" {
			return "UE List Tool v1.0.0"
		}
		return "Connected UEs:\n" + strings.Join(s.activeUEs[:2], "\n")

	default:
		return fmt.Sprintf("Executing %s command for gNodeB %s", command, gnbName)
	}
}

func (s *Server) handleUeCommand(cmd string, nodeName string) string {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return "Empty command"
	}

	commandName := parts[0]
	args := parts[1:]

	if command, exists := s.ueCommands[commandName]; exists {
		// Check for --help flag first
		if len(args) > 0 && args[0] == "--help" {
			if helpText, exists := command.SubCommands["--help"]; exists {
				return helpText // Return the help text directly from SubCommands
			}
		}
		return command.Handler(args, nodeName)
	}

	return fmt.Sprintf("Unknown command: %s", commandName)
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	server := NewServer()

	// Setup routes
	router.GET("/connect", server.handleConnect)
	router.GET("/dump", server.handleDump)
	router.POST("/command", server.handleCommand)

	log.Printf("Server starting on port 4000...")
	log.Fatal(router.Run(":4000"))
}
