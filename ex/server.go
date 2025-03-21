package ex

import (
	"client-server/models"

	"github.com/abiosoft/ishell/v2"
)

type NodeType int

const (
	UE NodeType = iota
	Gnb
	Amf
)

type Shell struct {
	Nodes []Node
	Ip    string
	Port  int
	Shell *ishell.Shell
}

type Node struct {
	Type        NodeType
	activeNodes []string
	Command     []models.Command
	Shell       *ishell.Shell
}

// WARN: just example for input params of handler func
type FormArgs struct {
	//????????
}

func (s *Shell) SetupShellUE(fns []func(any)) {
	for _, ue := range s.Nodes {
		if ue.Type == UE {
			for i := range ue.Command {
				//TODO: setup Command vs args
				//TODO: setup form -> fn
				form := FormArgs{}
				ue.Command[i].Fn = func(s []models.Subcommand) {
					fns[i](form)
				}
			}
		}
	}
}

//TODO: add http for listening req from client
