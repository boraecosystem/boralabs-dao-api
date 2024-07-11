package v1

import (
	"github.com/gin-gonic/gin"
)

type RESTInterface interface {
	routes(group *gin.RouterGroup)
	create(ctx *gin.Context)
	find(ctx *gin.Context)
	findAll(ctx *gin.Context)
}

type REST struct{}

// RoutesV1 REST API Version 1
func (r REST) RoutesV1(g *gin.RouterGroup) {
	ProposalV1{}.routes(g) // proposal and vote
}
