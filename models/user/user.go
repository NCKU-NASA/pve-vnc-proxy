package user

import (
//    "log"
//    "fmt"
//    "time"
//    "strings"
    "bytes"
    "io/ioutil"
    "encoding/json"
    "github.com/tidwall/gjson"
//    "io/ioutil"

//    "github.com/go-errors/errors"
//    "github.com/spf13/viper"
    "github.com/gin-gonic/gin"
//    "golang.org/x/crypto/ssh/terminal"

    "pve-vnc-proxy/utils/backend"
    "pve-vnc-proxy/middlewares/sessions"
//    "pve-vnc-proxy/utils/password"
//    "pve-vnc-proxy/models/group"
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

/*func GetUsers() ([]*User, error) {
    rows, err := database.Query(fmt.Sprintf("SELECT * FROM %s", tablename))
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var users []*User
    for rows.Next() {
        result := new(User)
        rows.Scan(&result.Username, &result.Password, &result.Online, &result.Enable, &result.Startdate, &result.Enddate)
        users = append(users, result)
    }
    return users, nil
}

func AddUser(user *User) error {
    testuser, err := GetUser(user.Username)
    if err != nil {
        return err
    }

    if testuser != nil {
        return errors.New("User is exist")
    }

    _, err = database.Exec(fmt.Sprintf("INSERT INTO %s (username, password, enable, startdate, enddate) VALUES (?,?,?,?,?)", tablename), user.Username, user.Password.String(), user.Enable, user.Startdate, user.Enddate)
    return err
}

func DeleteUser(username string) error {
    testuser, err := GetUser(username)
    if err != nil {
        return err
    }

    if testuser == nil {
        return errors.New("User not exist")
    }
    
    _, err = database.Exec(fmt.Sprintf("DELETE FROM %s where username=?", tablename), username)
    return err
}

func UpdateUser(username string, user *User, fields ...string) error {
    testuser, err := GetUser(username)
    if err != nil {
        return err
    }

    if testuser == nil {
        return errors.New("User not exist")
    }

    usermap := user.ToMap()

    data := []any{}
    for index := 0; index < len(fields); index++ {
        if fields[index] == "Username" {
            fields = append(fields[:index], fields[index+1:]...)
            index--
        } else {
            data = append(data, usermap[fields[index]])
        }
    }
    data = append(data, username)

    if len(fields) <= 0 {
        return nil
    }
    _, err = database.Exec(fmt.Sprintf("UPDATE %s SET %s where username=?", tablename, strings.ToLower(strings.Join(fields[:], "=?, ")) + "=?"), data...)
    return err
}*/
