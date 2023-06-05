package router
import (
    "github.com/gin-gonic/gin"

    "pve-vnc-proxy/router/vm"
    "pve-vnc-proxy/router/novnc"
)

var router *gin.RouterGroup

func Init(r *gin.RouterGroup) {
    router = r
    vm.Init(router.Group("/vm"))
    novnc.Init(router.Group("/novnc"))
}

