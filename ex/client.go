package ex

import "github.com/abiosoft/ishell/v2"

func (s *Shell) newShell(id NodeType) {
	s = &Shell{
		Nodes: []Node{
			Node{
				Type: id,
				//TODO: get all active nodes
				Shell: ishell.New(),
			},
		},
	}
}

func Connect(s *Shell, nodeType NodeType) {
	//TODO: connect to server ( -p ip:port)
	s.newShell(nodeType)
	return
}

// client: just create shell; send cmd (for ex FormArgs) to server every user excute terminal cmd line.
// server:
//	- send client: info nodes + args
//	- abstract for ue/gnb controller:
//		- UE can add excute func to `server`
