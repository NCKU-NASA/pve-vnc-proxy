package pve

import (
    "fmt"
    "io"
    "io/ioutil"
    "strings"
    "crypto/tls"
    "github.com/tidwall/gjson"
    "net/http"
    "net/url"
    "github.com/spf13/viper"
    "github.com/gorilla/websocket"
)

type loginticket struct {
    logintype string
    ticket string
    csrftoken string
}

func init() {
    http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
}

func login(username, vmname string) (loginticket, error) {
    nodename := viper.GetString(fmt.Sprintf("users.%s.%s.node", username, vmname))
    pveusername := viper.GetString(fmt.Sprintf("nodes.%s.username", nodename))
    pvetoken := viper.GetString(fmt.Sprintf("nodes.%s.token", nodename))
    if pvetoken != "" {
        return loginticket{
            logintype: "token",
            ticket: pveusername+"="+pvetoken,
        }, nil
    }
    endpoint := viper.GetString(fmt.Sprintf("nodes.%s.endpoint", nodename))
    pvepassword := viper.GetString(fmt.Sprintf("nodes.%s.password", nodename))
    data := url.Values(map[string][]string{
        "username": []string{pveusername},
        "password": []string{pvepassword},
    })
    useurl, _ := url.JoinPath("https://", endpoint, "/api2/json/access/ticket")
    req, err := http.NewRequest("POST", useurl, strings.NewReader(data.Encode()))
    if err != nil {
        return loginticket{}, err
    }
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    res, err := Start(req)
    if err != nil {
        return loginticket{}, err
    }
    defer res.Body.Close()
    result, err := ioutil.ReadAll(res.Body)
    if err != nil {
        return loginticket{}, err
    }
    return loginticket{
        logintype: "password",
        ticket: gjson.Get(string(result), "data.ticket").String(),
        csrftoken: gjson.Get(string(result), "data.CSRFPreventionToken").String(),
    }, nil
}

func genreq(username, vmname, protocol, method, uri string, body io.Reader) (*http.Request, loginticket, error) {
    ticket, err := login(username, vmname)
    if err != nil {
        return nil, ticket, err
    }
    nodename := viper.GetString(fmt.Sprintf("users.%s.%s.node", username, vmname))
    endpoint := viper.GetString(fmt.Sprintf("nodes.%s.endpoint", nodename))
    useurl, _ := url.JoinPath(protocol, endpoint, "/", uri)
    req, err := http.NewRequest(method, useurl, body)
    if err != nil {
        return nil, ticket, err
    }
    if ticket.logintype == "token" {
        req.Header.Set("Authorization", fmt.Sprintf("PVEAPIToken=%s", ticket.ticket))
    } else if ticket.logintype == "password" {
        req.Header.Set("CSRFPreventionToken", ticket.csrftoken)
        req.AddCookie(&http.Cookie{
            Name: "PVEAuthCookie",
            Value: ticket.ticket,
        })
    }
    return req, ticket, nil
}

func NewRequest(username, vmname, method, uri string, body io.Reader) (*http.Request, loginticket, error) {
    return genreq(username, vmname, "https://", method, uri, body)
}

func GetWebsocket(username, vmname, uri string, query url.Values) (*websocket.Conn, error) {
    req, _, err := genreq(username, vmname, "wss://", "GET", uri, nil)
    if err != nil {
        return nil, err
    }
    if query != nil {
        query.Del("su")
        query.Del("vmname")
        req.URL.RawQuery = query.Encode()
    }
    dialer := websocket.Dialer{
        Subprotocols: []string{"binary"},
    }
//    fmt.Println(req.URL.String())
//    fmt.Println(req.Cookies())
//    fmt.Println(req.Trailer)
    connect, _, err := dialer.Dial(req.URL.String(), req.Header)
    //fmt.Println(res)
    return connect, err
}

func Start(req *http.Request) (*http.Response, error) {
    client := &http.Client{}
    return client.Do(req)
}

func ClearInfo(ticket loginticket, data []byte) []byte {
    s := string(data)
    s = strings.Replace(s, ticket.ticket, "", -1)
    s = strings.Replace(s, ticket.csrftoken, "", -1)
    return []byte(s)
}
