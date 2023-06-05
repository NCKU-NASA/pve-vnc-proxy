package router
import (
    "github.com/gin-gonic/gin"
//    "net/http"

    "pve-vnc-proxy/router/vm"
    "pve-vnc-proxy/router/novnc"
//    "pve-vnc-proxy/models/user"
//    "pve-vnc-proxy/router/vpn"
    "pve-vnc-proxy/middlewares/auth"
    "pve-vnc-proxy/middlewares/sessions"
//    "pve-vnc-proxy/utils/error"
)

var router *gin.RouterGroup

func Init(r *gin.RouterGroup) {
    router = r
    router.GET("/status", auth.CheckSignIn, status)
    router.GET("/set", auth.CheckSignIn, set)
    vm.Init(router.Group("/vm"))
    novnc.Init(router.Group("/novnc"))
//    group.Init(router.Group("/group"))
//    vpn.Init(router.Group("/vpn"))
}

func set(c *gin.Context) {
    session := sessions.Default(c)
    session.Set("test", 666)
    c.JSON(200, true)
}

func status(c *gin.Context) {
    session := sessions.Default(c)
//    username := session.Get("user.username").String()
//    panic("dead")
    /*error.AbortAndError(c, &error.Err{
        Code: 401,
        Msg: "test bad",
    })
    c.String(200, "test2")*/
//    userdata := user.GetUser(c, username)
    c.JSON(200, session.GetJSON())
}
