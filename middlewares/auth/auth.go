package auth

import (
    "github.com/gin-gonic/gin"

    "pve-vnc-proxy/utils/errutil"
    "pve-vnc-proxy/models/user"
    "pve-vnc-proxy/middlewares/sessions"
)

func CheckSignIn(c *gin.Context) {
    if isSignIn, exist := c.Get("isSignIn"); !exist || !isSignIn.(bool) {
        errutil.AbortAndStatus(c, 401)
    }
}

func CheckIsAdmin(c *gin.Context) {
    if isAdmin, exist := c.Get("isAdmin"); !exist || !isAdmin.(bool) {
        errutil.AbortAndStatus(c, 401)
    }
}

func SetNoVNCSession(fullreplace bool) gin.HandlerFunc {
    return func(c *gin.Context) {
        session := sessions.Default(c)
        data := make(map[string]any)
        if !fullreplace {
            for k, v := range session.Get("novnc").Map() {
                data[k] = v.String()
            }
        }
        query := make(map[string]any)
        err := c.ShouldBind(&query)
        if err != nil {
            query := make(map[string][]string)
            c.Bind(&query)
            for k, v := range query {
                if v[0] != "null" {
                    data[k] = v[0]
                }
            }
        } else {
            for k, v := range query {
                if v != "null" {
                    data[k] = v
                }
            }
        }
        session.Set("novnc", data)
    }
}

func AddMeta(c *gin.Context) {
    session := sessions.Default(c)
    username := session.Get("user.username").String()
    if username == "" {
        c.Set("isSignIn", false)
    } else {
        userdata := user.GetUser(c, username)
        if userdata == nil {
            c.Set("isSignIn", false)
        } else {
            c.Set("isSignIn", true)
            c.Set("isAdmin", userdata.IsAdmin)
        }
    }
}
