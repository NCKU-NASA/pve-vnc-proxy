package novnc

import (
//    "log"
    "fmt"
    "net/url"

    "github.com/gin-gonic/gin"
    "github.com/spf13/viper"
//    "github.com/gin-contrib/sessions"

    "pve-vnc-proxy/middlewares/auth"
//    "pve-vnc-proxy/middlewares/sessions"
//    "pve-vnc-proxy/models/user"
    "pve-vnc-proxy/models/pve"
//    "pve-vnc-proxy/utils/errutil"
//    "pve-vnc-proxy/utils/password"
//    "net/http"
)

var router *gin.RouterGroup

func Init(r *gin.RouterGroup) {
    router = r
    router.GET("/app.js", auth.SetNoVNCSession(false), auth.CheckSignIn, getapp)
    router.GET("/package.json", auth.SetNoVNCSession(false), auth.CheckSignIn, getpackage)
    router.GET("/app/*path", auth.SetNoVNCSession(false), auth.CheckSignIn, getdata)
//    router.POST("/getuser", auth.CheckSignIn, auth.CheckIsAdmin, userinfo)
//    router.GET("/test", auth.CheckSignIn, auth.CheckIsAdmin, test)
//    router.POST("/test", auth.CheckSignIn, auth.CheckIsAdmin, test)
//    router.GET("/getall", auth.CheckSignIn, auth.CheckIsAdmin, alluserinfo)
//    router.POST("/login", login)
//    router.POST("/add", auth.CheckSignIn, auth.CheckIsAdmin, adduser)
//    router.GET("/logout", logout)
//    router.GET("/issignin", issignin)
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
