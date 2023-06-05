package vm

import (
    "log"
    "fmt"
    "net/url"
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/gorilla/websocket"
    "github.com/spf13/viper"
    "github.com/tidwall/gjson"

    "pve-vnc-proxy/middlewares/auth"
    "pve-vnc-proxy/middlewares/sessions"
    "pve-vnc-proxy/models/pve"
    "pve-vnc-proxy/utils/errutil"
)

var router *gin.RouterGroup

func Init(r *gin.RouterGroup) {
    router = r
    router.GET("", auth.SetNoVNCSession(true), auth.CheckSignIn, showvm)
    router.GET("/api/status/:cmd", auth.SetNoVNCSession(false), auth.CheckSignIn, command)
    router.POST("/api/status/:cmd", auth.SetNoVNCSession(false), auth.CheckSignIn, command)
    router.POST("/api/vncproxy", auth.SetNoVNCSession(false), auth.CheckSignIn, vncproxy)
    router.GET("/api/vncwebsocket", auth.SetNoVNCSession(false), auth.CheckSignIn, vncwebsocket)
}

func showvm(c *gin.Context) {
    session := sessions.Default(c)
    username, vmname := pve.LoginInfo(c)
    if session.Get("user.username").String() == username {
        pve.Config(c, username, vmname)
    }
    c.Data(pve.ShowVM(username, vmname))
}

func command(c *gin.Context) {
    username, vmname := pve.LoginInfo(c)
    node := viper.GetString(fmt.Sprintf("users.%s.%s.node", username, vmname))
    node = viper.GetString(fmt.Sprintf("nodes.%s.node", node))
    vmid := viper.GetString(fmt.Sprintf("users.%s.%s.vmid", username, vmname))
    useuri, _ := url.JoinPath("/api2/json/nodes", node, "qemu", vmid, "status", c.Param("cmd"))
    c.Data(pve.Proxy(username, vmname, c.Request.Method, useuri))
}

func vncproxy(c *gin.Context) {
    username, vmname := pve.LoginInfo(c)
    data, err := pve.VNCProxy(username, vmname)
    if err != nil {
        errutil.AbortAndStatus(c, 404)
        return
    }
    result := map[string]map[string]string{
        "data": map[string]string{
            "port": gjson.Get(string(data), "data.port").String(),
            "ticket": gjson.Get(string(data), "data.ticket").String(),
        },
    }
    c.JSON(200, result)
}

func vncwebsocket(c *gin.Context) {
    username, vmname := pve.LoginInfo(c)

    upgrader := websocket.Upgrader{
        Subprotocols: []string{"binary"},
        CheckOrigin: func(r *http.Request) bool { 
            return true 
        },
    }
    ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        log.Panicln(err)
    }
    pve.Tunnel(username, vmname, c.Request.URL.Query(), ws)
}

