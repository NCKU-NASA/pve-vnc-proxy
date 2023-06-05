package sessions
import (
//    "fmt"
    "bytes"
    "io/ioutil"
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/tidwall/gjson"
    "github.com/tidwall/sjson"
    
    "pve-vnc-proxy/utils/backend"
)

type sessions struct {
    sessionscookie *http.Cookie
    session string
}

func Sessions(name string) gin.HandlerFunc {
    return func(c *gin.Context) {
        sessions := &sessions{
            sessionscookie: nil,
            session: "",
        }
        c.Set("sessions", sessions)
        cookie, err := c.Cookie(name)
        if err != nil {
            return
        }
        sessions.sessionscookie = &http.Cookie{
            Name: name,
            Value: cookie,
        }

        req, err := backend.NewRequest(c, "GET", "sessions/get", nil)
        if err != nil {
            return
        }
        res, err := backend.Start(req)
        if err != nil {
            return
        }
        defer res.Body.Close()
        result, err := ioutil.ReadAll(res.Body)
        if err != nil {
            return
        }
        sessions.session = string(result)

        c.Next()

        if sessions.session != string(result) {
            req, err = backend.NewRequest(c, "POST", "sessions/set", bytes.NewBuffer([]byte(sessions.session)))
            req.Header.Set("Content-Type", "application/json")
            if err != nil {
                return
            }
            res, err = backend.Start(req)
            if err != nil {
                return
            }
            defer res.Body.Close()
        }
    }
}

func Default(c *gin.Context) *sessions {
    result, _ := c.Get("sessions")
    return result.(*sessions)
}

func (s *sessions) GetSessionCookie() *http.Cookie {
    tmp := *s.sessionscookie
    return &tmp
}

func (s *sessions) GetJSON() string {
    return s.session
}

func (s *sessions) Get(name string) gjson.Result {
    return gjson.Get(s.session, name)
}

func (s *sessions) Set(name string, value any) {
    s.session, _ = sjson.Set(s.session, name, value)
}
