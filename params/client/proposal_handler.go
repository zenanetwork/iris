package client

import (
	govclient "github.com/zenanetwork/iris/gov/client"
	"github.com/zenanetwork/iris/params/client/cli"
	"github.com/zenanetwork/iris/params/client/rest"
)

// param change proposal handler
var ProposalHandler = govclient.NewProposalHandler(cli.GetCmdSubmitProposal, rest.ProposalRESTHandler)
