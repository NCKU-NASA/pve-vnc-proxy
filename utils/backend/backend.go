package backend

import (
    "io"
    "net/http"
    "crypto/tls"
    "github.com/gin-gonic/gin"
    "github.com/spf13/viper"
//    "github.com/go-errors/errors"
//    "pve-vnc-proxy/middlewares/sessions"
)

type sessions interface {
    GetSessionCookie() *http.Cookie
}

func init() {
    http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
}

func NewRequest(c *gin.Context, method, uri string, body io.Reader) (*http.Request, error) {
    session, _ := c.Get("sessions")
    req, err := http.NewRequest(method, viper.GetString("BackendURL")+"/"+uri, body)
    if err != nil {
        return nil, err
    }
    if session.(sessions).GetSessionCookie() != nil {
        req.AddCookie(session.(sessions).GetSessionCookie())
    }
    return req, nil
}

func Start(req *http.Request) (*http.Response, error) {
    client := &http.Client{}
    return client.Do(req)
}
