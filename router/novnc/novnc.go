package novnc

import (
    "fmt"
    "net/url"

    "github.com/gin-gonic/gin"
    "github.com/spf13/viper"

    "pve-vnc-proxy/middlewares/auth"
    "pve-vnc-proxy/models/pve"
)

var router *gin.RouterGroup

func Init(r *gin.RouterGroup) {
    router = r
    router.GET("/app.js", auth.SetNoVNCSession(false), auth.CheckSignIn, getapp)
    router.GET("/package.json", auth.SetNoVNCSession(false), auth.CheckSignIn, getpackage)
    router.GET("/app/*path", auth.SetNoVNCSession(false), auth.CheckSignIn, getdata)
}

func getapp(c *gin.Context) {
    username, vmname := pve.LoginInfo(c)
    node := viper.GetString(fmt.Sprintf("users.%s.%s.node", username, vmname))
    app := viper.GetString(fmt.Sprintf("nodes.%s.app", node))
    path, _ := url.JoinPath("app/", app)
    c.File(path)
}

func getpackage(c *gin.Context) {
    username, vmname := pve.LoginInfo(c)
    c.Data(pve.Proxy(username, vmname, "GET", "/novnc/package.json"))
}

func getdata(c *gin.Context) {
    username, vmname := pve.LoginInfo(c)
    useuri, _ := url.JoinPath("/novnc/app/", c.Param("path"))
    c.Data(pve.Proxy(username, vmname, "GET", useuri))
}
