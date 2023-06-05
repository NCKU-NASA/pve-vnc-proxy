package user

import (
    "bytes"
    "io/ioutil"
    "encoding/json"
    "github.com/tidwall/gjson"

    "github.com/gin-gonic/gin"

    "pve-vnc-proxy/utils/backend"
    "pve-vnc-proxy/middlewares/sessions"
)
type user struct {
    Username string
    StudentId string
    Password string
    IPIndex int64
    IsAdmin bool
}

func (user *user) ToMap() map[string]any {
    var usermap map[string]any
    userjson, _ := json.Marshal(user)
    json.Unmarshal(userjson, &usermap)
    return usermap
}

func GetUser(c *gin.Context, username string) *user {
    session := sessions.Default(c)
    data, _ := json.Marshal(map[string]string{
        "username": username,
    })
    req, err := backend.NewRequest(c, "POST", "user/userdata", bytes.NewBuffer(data))
    if err != nil {
        return nil
    }
    req.Header.Set("Content-Type", "application/json")
    res, err := backend.Start(req)
    if err != nil {
        return nil
    }
    defer res.Body.Close()
    result, err := ioutil.ReadAll(res.Body)
    if err != nil {
        return nil
    }
    output := new(user)
    output.Username = gjson.Get(string(result), "username").String()
    output.StudentId = gjson.Get(string(result), "studentId").String()
    if session.Get("user.username").String() == output.Username {
        output.Password = session.Get("user.password").String()
    }
    output.IPIndex = gjson.Get(string(result), "ipindex").Int()
    output.IsAdmin = false
    for _, v := range gjson.Get(string(result), "groups").Array() {
        if v.String() == "admin" {
            output.IsAdmin = true
        }
    }
    return output
}

func GetSSHKey(c *gin.Context) string {
    req, err := backend.NewRequest(c, "GET", "pubkey", nil)
    if err != nil {
        return ""
    }
    res, err := backend.Start(req)
    if err != nil {
        return ""
    }
    defer res.Body.Close()
    result, err := ioutil.ReadAll(res.Body)
    if err != nil {
        return ""
    }
    return string(result)
}
