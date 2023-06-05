package main
import (
//    "net/http"
    "github.com/gin-gonic/gin"
//    "github.com/gin-contrib/sessions"
//    "github.com/gin-contrib/sessions/cookie"
    "github.com/go-errors/errors"
    "github.com/spf13/viper"

    "pve-vnc-proxy/router"
    _ "pve-vnc-proxy/utils/config"
//    "pve-vnc-proxy/utils/database"
    "pve-vnc-proxy/utils/errutil"
//    _ "pve-vnc-proxy/models/user"
    "pve-vnc-proxy/middlewares/auth"
    "pve-vnc-proxy/middlewares/sessions"
)

func main() {
//    defer database.Close()
//    store := cookie.NewStore([]byte(config.Secret))
    if !viper.GetBool("debug") {
        gin.SetMode(gin.ReleaseMode)
    }
    backend := gin.Default()
    backend.Use(errorHandler)
    backend.Use(gin.CustomRecovery(panicHandler))
    backend.Use(sessions.Sessions(viper.GetString("SessionsName")))
    backend.Use(auth.AddMeta)
    router.Init(&backend.RouterGroup)
    backend.Run(":"+viper.GetString("Port"))
}

func panicHandler(c *gin.Context, err any) {
    goErr := errors.Wrap(err, 2)
    data := ""
    if viper.GetBool("debug") {
        data = goErr.Error()
    }
    errutil.AbortAndError(c, &errutil.Err{
        Code: 500,
        Msg: "Internal server error",
        Data: data,
    })
}

func errorHandler(c *gin.Context) {
    c.Next()

    for _, e := range c.Errors {
        err := e.Err
        if myErr, ok := err.(*errutil.Err); ok {
            if myErr.Msg != nil {
                c.JSON(myErr.Code, myErr.ToH())
            } else {
                c.Status(myErr.Code)
            }
        } else {
            c.JSON(500, gin.H{
                "code": 500,
                "msg": "Internal server error",
                //"data": err.Error(),
            })
        }
        return
    }
}
