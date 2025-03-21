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

// IShell interface for command execution
type IShell interface {
	LoadCommands(config models.CommandConfig)
	SetupCommands(shell *ishell.Shell, nodeName string)
	HandleAPIRequest(command string, nodeName string) string
}

// UEShell implements IShell for UE commands
type UEShell struct {
	commands []models.Command
	server   *Server
}

// GnbShell implements IShell for gNodeB commands
type GnbShell struct {
	commands []models.Command
	server   *Server
}

type Server struct {
	activeUEs    []string
	gNodeBs      []string
	commandsConf models.CommandConfig
	ueShell      IShell
	gnbShell     IShell
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

	s.ueShell = &UEShell{server: s}
	s.gnbShell = &GnbShell{server: s}

	return s
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

	s.ueShell.LoadCommands(s.commandsConf)
	s.gnbShell.LoadCommands(s.commandsConf)

	return nil
}

// Render a response by replacing variables
func renderResponse(response string, nodeName string, args []string) string {
	result := strings.Replace(response, "${nodeName}", nodeName, -1)

	for i, arg := range args {
		if i > 0 { // Skip the subcommand itself
			argPlaceholder := fmt.Sprintf("${arg%d}", i)
			result = strings.Replace(result, argPlaceholder, arg, -1)
		}
	}

	return result
}

// Load UE commands
func (u *UEShell) LoadCommands(config models.CommandConfig) {
	u.commands = config.UE.Commands
}

// Setup UE commands using ishell's command tree
func (u *UEShell) SetupCommands(shell *ishell.Shell, nodeName string) {
	for _, cmdConfig := range u.commands {
		// Create the parent command
		mainCmd := &ishell.Cmd{
			Name: cmdConfig.Name,
			Help: cmdConfig.Help,
			Func: func(c *ishell.Context) {
				// If no args, show usage
				if len(c.Args) == 0 {
					c.Println(cmdConfig.Usage)
					return
				}

				// Try to execute as a subcommand
				cmd, args := c.Cmd.FindCmd(c.Args)
				if cmd == nil {
					c.Println("Unknown subcommand:", c.Args[0])
					c.Println("Available subcommands:")
					for _, subcmd := range c.Cmd.Children() {
						c.Printf("  %s: %s\n", subcmd.Name, subcmd.Help)
					}
					return
				}

				// Command found, execute it
				cmd.Func(ishell.NewContext(shell, cmd, args))
			},
		}

		// Add subcommands
		for _, subConfig := range cmdConfig.Subcommands {
			subCmd := &ishell.Cmd{
				Name: subConfig.Name,
				Help: subConfig.Help,
				Func: func(subConfig models.Subcommand) func(c *ishell.Context) {
					return func(c *ishell.Context) {
						response := renderResponse(subConfig.Response, nodeName, c.Args)
						c.Println(response)
					}
				}(subConfig),
			}
			mainCmd.AddCmd(subCmd)
		}

		// Add the main command to shell
		shell.AddCmd(mainCmd)
	}
}

// Handle API requests for UE commands
func (u *UEShell) HandleAPIRequest(command string, nodeName string) string {
	// Create a temporary shell to process the command
	shell := ishell.New()

	// Set up commands in the shell
	u.SetupCommands(shell, nodeName)

	// Capture output
	var outputBuffer strings.Builder
	shell.SetOut(&outputBuffer)

	// Process the command
	cmdParts := strings.Fields(command)
	if len(cmdParts) == 0 {
		return "Empty command"
	}

	err := shell.Process(cmdParts...)
	if err != nil {
		return fmt.Sprintf("Error executing command: %v", err)
	}

	return outputBuffer.String()
}

// Load Gnb commands
func (g *GnbShell) LoadCommands(config models.CommandConfig) {
	g.commands = config.GNB.Commands
}

// Setup Gnb commands using ishell's command tree
func (g *GnbShell) SetupCommands(shell *ishell.Shell, nodeName string) {
	for _, cmdConfig := range g.commands {
		// Create the parent command
		mainCmd := &ishell.Cmd{
			Name: cmdConfig.Name,
			Help: cmdConfig.Help,
			Func: func(c *ishell.Context) {
				// If no args, show usage
				if len(c.Args) == 0 {
					c.Println(cmdConfig.Usage)
					return
				}

				// Try to execute as a subcommand
				cmd, args := c.Cmd.FindCmd(c.Args)
				if cmd == nil {
					c.Println("Unknown subcommand:", c.Args[0])
					c.Println("Available subcommands:")
					for _, subcmd := range c.Cmd.Children() {
						c.Printf("  %s: %s\n", subcmd.Name, subcmd.Help)
					}
					return
				}

				// Command found, execute it
				cmd.Func(ishell.NewContext(shell, cmd, args))
			},
		}

		// Add subcommands
		for _, subConfig := range cmdConfig.Subcommands {
			subCmd := &ishell.Cmd{
				Name: subConfig.Name,
				Help: subConfig.Help,
				Func: func(subConfig models.Subcommand) func(c *ishell.Context) {
					return func(c *ishell.Context) {
						response := renderResponse(subConfig.Response, nodeName, c.Args)
						c.Println(response)
					}
				}(subConfig),
			}
			mainCmd.AddCmd(subCmd)
		}

		// Add the main command to shell
		shell.AddCmd(mainCmd)
	}
}

// Handle API requests for Gnb commands
func (g *GnbShell) HandleAPIRequest(command string, nodeName string) string {
	// Create a temporary shell to process the command
	shell := ishell.New()

	// Set up commands in the shell
	g.SetupCommands(shell, nodeName)

	// Capture output
	var outputBuffer strings.Builder
	shell.SetOut(&outputBuffer)

	// Process the command
	cmdParts := strings.Fields(command)
	if len(cmdParts) == 0 {
		return "Empty command"
	}

	err := shell.Process(cmdParts...)
	if err != nil {
		return fmt.Sprintf("Error executing command: %v", err)
	}

	return outputBuffer.String()
}

// API Handlers
func (s *Server) handleConnect(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "Connected successfully",
		"objects": gin.H{
			"ue":  "User Equipment",
			"gnb": "gNodeB",
		},
	})
}

func (s *Server) handleDump(c *gin.Context) {
	nodeType := c.Param("nodeType")

	if nodeType == "ue" {
		c.JSON(200, gin.H{"nodes": s.activeUEs})
	} else if nodeType == "gnb" {
		c.JSON(200, gin.H{"nodes": s.gNodeBs})
	} else {
		c.JSON(400, gin.H{"error": "Invalid node type"})
	}

}

func (s *Server) handleCommand(c *gin.Context) {
	//dùng ishell structure response để xử lý command.
	var request struct {
		Command  string `json:"command" binding:"required"`
		NodeType string `json:"nodeType" binding:"required"`
		NodeName string `json:"nodeName" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var response string
	if request.NodeType == "ue" {
		response = s.ueShell.HandleAPIRequest(request.Command, request.NodeName)
	} else if request.NodeType == "gnb" {
		response = s.gnbShell.HandleAPIRequest(request.Command, request.NodeName)
	} else {
		c.JSON(400, gin.H{"error": "Invalid node type"})
		return
	}

	c.JSON(200, gin.H{"response": response})

}

func (s *Server) handleGetCommands(c *gin.Context) {
	nodeType := c.Param("nodeType")

	var commands []map[string]string

	if nodeType == "ue" {
		for _, cmd := range s.commandsConf.UE.Commands {
			commands = append(commands, map[string]string{
				"name":  cmd.Name,
				"help":  cmd.Help,
				"usage": cmd.Usage,
			})
		}
	} else if nodeType == "gnb" {
		for _, cmd := range s.commandsConf.GNB.Commands {
			commands = append(commands, map[string]string{
				"name":  cmd.Name,
				"help":  cmd.Help,
				"usage": cmd.Usage,
			})
		}
	} else {
		c.JSON(400, gin.H{"error": "Invalid node type"})
		return
	}

	c.JSON(200, gin.H{"commands": commands})
}

// Check if node exists
func (s *Server) handleCheckNode(c *gin.Context) {
	nodeType := c.Param("nodeType")
	nodeName := c.Param("nodeName")

	exists := false

	if nodeType == "ue" {
		for _, name := range s.activeUEs {
			if name == nodeName {
				exists = true
				break
			}
		}
	} else if nodeType == "gnb" {
		for _, name := range s.gNodeBs {
			if name == nodeName {
				exists = true
				break
			}
		}
	} else {
		c.JSON(400, gin.H{"error": "Invalid node type"})
		return
	}

	c.JSON(200, gin.H{"exists": exists})
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
	router.GET("/dump/:nodeType", server.handleDump)
	router.POST("/command", server.handleCommand)
	router.GET("/commands/:nodeType", server.handleGetCommands)
	router.GET("/check/:nodeType/:nodeName", server.handleCheckNode)

	// Start the HTTP server
	log.Printf("Server starting on port 4000...")
	log.Fatal(router.Run(":4000"))
}
