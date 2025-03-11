package main

import (
	"client-server/models"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/abiosoft/ishell/v2"
	"github.com/gin-gonic/gin"
)

var (
	configFile = flag.String("c", "command.json", "Path to the configuration file")
)

type Server struct {
	activeUEs    []string
	gNodeBs      []string
	commandsConf models.CommandConfig
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

// LoadCommandsConfig loads command configurations from a JSON file
func (s *Server) LoadCommandsConfig(filePath string) error {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading config file: %v", err)
	}

	err = json.Unmarshal(data, &s.commandsConf)
	if err != nil {
		return fmt.Errorf("error parsing config file: %v", err)
	}

	return nil
}

// Render a response by replacing variables
func renderResponse(response string, nodeName string, args []string) string {
	result := strings.Replace(response, "${nodeName}", nodeName, -1)

	// Replace ${arg1}, ${arg2}, etc. with actual arguments if provided
	for i, arg := range args {
		if i > 0 { // Skip the subcommand itself
			argPlaceholder := fmt.Sprintf("${arg%d}", i)
			result = strings.Replace(result, argPlaceholder, arg, -1)
		}
	}

	return result
}

// Setup UE commands
func (s *Server) setupUECommands(shell *ishell.Shell, ueName string) {
	for _, cmd := range s.commandsConf.UE.Commands {
		// Create a copy of cmd for the closure
		cmdCopy := cmd

		shell.AddCmd(&ishell.Cmd{
			Name: cmdCopy.Name,
			Help: cmdCopy.Help,
			Func: func(c *ishell.Context) {
				if len(c.Args) == 0 {
					c.Println(cmdCopy.DefaultUsage)
					return
				}

				subcommand := c.Args[0]
				found := false

				for _, sub := range cmdCopy.Subcommands {
					if sub.Name == subcommand {
						found = true
						c.Println(renderResponse(sub.Response, ueName, c.Args))
						break
					}
				}

				if !found {
					// Check if there's a default handler
					for _, sub := range cmdCopy.Subcommands {
						if sub.Name == "default" {
							c.Println(renderResponse(sub.Response, ueName, c.Args))
							return
						}
					}
					c.Println("Invalid subcommand for " + cmdCopy.Name)
				}
			},
		})
	}
}

// Setup gNodeB commands
func (s *Server) setupGnbCommands(shell *ishell.Shell, gnbName string) {
	for _, cmd := range s.commandsConf.GNB.Commands {
		// Create a copy of cmd for the closure
		cmdCopy := cmd

		shell.AddCmd(&ishell.Cmd{
			Name: cmdCopy.Name,
			Help: cmdCopy.Help,
			Func: func(c *ishell.Context) {
				if len(c.Args) == 0 {
					c.Println(cmdCopy.DefaultUsage)
					return
				}

				subcommand := c.Args[0]
				found := false

				for _, sub := range cmdCopy.Subcommands {
					if sub.Name == subcommand {
						found = true
						c.Println(renderResponse(sub.Response, gnbName, c.Args))
						break
					}
				}

				if !found {
					// Check if there's a default handler
					for _, sub := range cmdCopy.Subcommands {
						if sub.Name == "default" {
							c.Println(renderResponse(sub.Response, gnbName, c.Args))
							return
						}
					}
					c.Println("Invalid subcommand for " + cmdCopy.Name)
				}
			},
		})
	}
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

func (s *Server) handleGetCommands(c *gin.Context) {
	nodeType := c.Param("nodeType")

	var commands []map[string]string

	if nodeType == "ue" {
		for _, cmd := range s.commandsConf.UE.Commands {
			commands = append(commands, map[string]string{
				"name":         cmd.Name,
				"help":         cmd.Help,
				"defaultUsage": cmd.DefaultUsage,
			})
		}
	} else if nodeType == "gnb" {
		for _, cmd := range s.commandsConf.GNB.Commands {
			commands = append(commands, map[string]string{
				"name":         cmd.Name,
				"help":         cmd.Help,
				"defaultUsage": cmd.DefaultUsage,
			})
		}
	} else {
		c.JSON(400, gin.H{"error": "Invalid node type"})
		return
	}

	c.JSON(200, gin.H{"commands": commands})
}

func main() {
	flag.Parse()
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	server := NewServer()

	// Load commands config
	if err := server.LoadCommandsConfig(*configFile); err != nil {
		log.Fatalf("Failed to load commands config: %v", err)
		os.Exit(1)
	}

	log.Printf("Commands loaded successfully from %s", *configFile)

	// Setup routes
	router.GET("/connect", server.handleConnect)
	router.GET("/dump", server.handleDump)
	router.POST("/command", server.handleCommand)
	router.GET("/commands/:nodeType", server.handleGetCommands)

	// Start the HTTP server
	log.Printf("Server starting on port 4000...")
	log.Fatal(router.Run(":4000"))
}
