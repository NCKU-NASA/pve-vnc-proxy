package vm

import (
    "log"
    "fmt"
    "net/url"
    "net/http"
//    "encoding/json"

    "github.com/gin-gonic/gin"
    "github.com/gorilla/websocket"
    "github.com/spf13/viper"
    "github.com/tidwall/gjson"
//    "github.com/gin-contrib/sessions"

    "pve-vnc-proxy/middlewares/auth"
//    "pve-vnc-proxy/middlewares/sessions"
//    "pve-vnc-proxy/models/user"
    "pve-vnc-proxy/models/pve"
    "pve-vnc-proxy/utils/errutil"
//    "pve-vnc-proxy/utils/password"
//    "net/http"
)

var router *gin.RouterGroup

func Init(r *gin.RouterGroup) {
    router = r
    router.GET("", auth.SetNoVNCSession(true), auth.CheckSignIn, showvm)
    router.GET("/api/status/:cmd", auth.SetNoVNCSession(false), auth.CheckSignIn, command)
    router.POST("/api/status/:cmd", auth.SetNoVNCSession(false), auth.CheckSignIn, command)
    router.POST("/api/vncproxy", auth.SetNoVNCSession(false), auth.CheckSignIn, vncproxy)
    router.GET("/api/vncwebsocket", auth.SetNoVNCSession(false), auth.CheckSignIn, vncwebsocket)
//    router.POST("/getuser", auth.CheckSignIn, auth.CheckIsAdmin, userinfo)
//    router.GET("/test", auth.CheckSignIn, auth.CheckIsAdmin, test)
//    router.POST("/test", auth.CheckSignIn, auth.CheckIsAdmin, test)
//    router.GET("/getall", auth.CheckSignIn, auth.CheckIsAdmin, alluserinfo)
//    router.POST("/login", login)
//    router.POST("/add", auth.CheckSignIn, auth.CheckIsAdmin, adduser)
//    router.GET("/logout", logout)
//    router.GET("/issignin", issignin)
}

func showvm(c *gin.Context) {
    username, vmname := pve.LoginInfo(c)
    pve.Config(c, username, vmname)
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
    //fmt.Println(ws.ReadMessage())
    pve.Tunnel(username, vmname, c.Request.URL.Query(), ws)
}

/*
func test(c *gin.Context) {
//    var data any
    data := make(map[string]any)
    err := c.ShouldBind(&data)
    if err != nil {
        data := make(map[string][]string)
        fmt.Println(c.ShouldBind(&data))
        fmt.Println(data)
        c.JSON(200, data)
        return
    }
    c.JSON(200, data)
}
func userinfo(c *gin.Context) {
    postdata := make(map[string]any)
    c.BindJSON(&postdata)
    userdata := user.GetUser(c, postdata["username"].(string))
    c.JSON(200, userdata)
}
*/
