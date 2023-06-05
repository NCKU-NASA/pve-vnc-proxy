package pve

import (
    "log"
    "fmt"
    "net"
    "io/ioutil"
    "strings"
    "net/url"
    "math/big"

    "github.com/spf13/viper"
    "github.com/gin-gonic/gin"
    "github.com/go-errors/errors"
    "github.com/gorilla/websocket"

    "pve-vnc-proxy/middlewares/sessions"
    "pve-vnc-proxy/utils/pve"
    "pve-vnc-proxy/utils/errutil"
    "pve-vnc-proxy/models/user"
)

func LoginInfo(c *gin.Context) (username, vmname string) {
    session := sessions.Default(c)
    username = session.Get("novnc.su").String()
    vmname = session.Get("novnc.vmname").String()
    if isAdmin, exist := c.Get("isAdmin"); username == "" || !exist || !isAdmin.(bool) {
        username = session.Get("user.username").String()
    }
    if vmname == "" {
        for k, _ := range viper.GetStringMap(fmt.Sprintf("users.%s", username)) {
            if k != "" {
                vmname = k
            }
        }
    }
    return
}

func ShowVM(username, vmname string) (int, string, []byte) {
    node := viper.GetString(fmt.Sprintf("users.%s.%s.node", username, vmname))
    vmid := viper.GetString(fmt.Sprintf("users.%s.%s.vmid", username, vmname))
    if node == "" {
        return 404, "text/plain", []byte("404 page not found")
    }
    req, ticket, err := pve.NewRequest(username, vmname, "GET", "/", nil)
    if err != nil {
        log.Panicln(err)
        return 500, "", []byte("")
    }
    data := req.URL.Query()
    data.Add("console", "kvm")
    data.Add("vmid", vmid)
    data.Add("node", node)
    data.Add("resize", "scale")
    data.Add("novnc", "1")
    req.URL.RawQuery = data.Encode()
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    res, err := pve.Start(req)
    if err != nil {
        log.Panicln(err)
        return 500, "", []byte("")
    }
    defer res.Body.Close()
    result, _ := ioutil.ReadAll(res.Body)
    return res.StatusCode, res.Header.Get("Content-Type"), pve.ClearInfo(ticket, result)
}

func Proxy(username, vmname, method, path string) (int, string, []byte) {
    node := viper.GetString(fmt.Sprintf("users.%s.%s.node", username, vmname))
    //vmid := viper.GetString(fmt.Sprintf("users.%s.%s.vmid", username, vmname))
    if node == "" {
        return 404, "text/plain", []byte("404 page not found")
    }
    req, ticket, err := pve.NewRequest(username, vmname, method, path, nil)
    if err != nil {
        log.Panicln(err)
        return 500, "", []byte("")
    }
    //fmt.Println(req.URL)
    res, err := pve.Start(req)
    if err != nil {
        log.Panicln(err)
        return 500, "", []byte("")
    }
    defer res.Body.Close()
    result, _ := ioutil.ReadAll(res.Body)
    return res.StatusCode, res.Header.Get("Content-Type"), pve.ClearInfo(ticket, result)
}

func VNCProxy(username, vmname string) ([]byte, error) {
    node := viper.GetString(fmt.Sprintf("users.%s.%s.node", username, vmname))
    vmid := viper.GetString(fmt.Sprintf("users.%s.%s.vmid", username, vmname))
    if node == "" {
        return nil, errors.New("404 page not found")
    }
    pvenode := viper.GetString(fmt.Sprintf("nodes.%s.node", node))
    useuri, _ := url.JoinPath("/api2/json/nodes", pvenode, "qemu", vmid, "vncproxy")
    data := url.Values(map[string][]string{
        "websocket": []string{"1"},
    })
    req, _, err := pve.NewRequest(username, vmname, "POST", useuri, strings.NewReader(data.Encode()))
    if err != nil {
        log.Panicln(err)
        return nil, errors.New("404 page not found")
    }
    //fmt.Println(req.URL)
    res, err := pve.Start(req)
    if err != nil {
        log.Panicln(err)
        return nil, errors.New("404 page not found")
    }
    defer res.Body.Close()
    result, _ := ioutil.ReadAll(res.Body)
    return result, nil
}

func Tunnel(username, vmname string, query url.Values, remote *websocket.Conn) {
    //fmt.Println(username, vmname)
    defer remote.Close()
    node := viper.GetString(fmt.Sprintf("users.%s.%s.node", username, vmname))
    vmid := viper.GetString(fmt.Sprintf("users.%s.%s.vmid", username, vmname))
    if node == "" {
        return
    }
    pvenode := viper.GetString(fmt.Sprintf("nodes.%s.node", node))
    useuri, _ := url.JoinPath("/api2/json/nodes", pvenode, "qemu", vmid, "vncwebsocket")

    pvews, err := pve.GetWebsocket(username, vmname, useuri, query)
    if err != nil {
        return
    }
    defer pvews.Close()
    go func() {
        for {
            messageType, data, err := remote.ReadMessage()
            if err != nil {
                return
            }
            err = pvews.WriteMessage(messageType, data)
            if err != nil {
                return
            }
        }
    }()
    for {
        messageType, data, err := pvews.ReadMessage()
        if err != nil {
            return
        }
        err = remote.WriteMessage(messageType, data)
        if err != nil {
            return
        }
    }
}

func Config(c *gin.Context, username, vmname string) {
    node := viper.GetString(fmt.Sprintf("users.%s.%s.node", username, vmname))
    vmid := viper.GetString(fmt.Sprintf("users.%s.%s.vmid", username, vmname))
    if node == "" {
        errutil.AbortAndStatus(c, 404)
        return
    }
    pvenode := viper.GetString(fmt.Sprintf("nodes.%s.node", node))
    useuri, _ := url.JoinPath("/api2/json/nodes", pvenode, "qemu", vmid, "config")
    userdata := user.GetUser(c, username)
    datamap := map[string][]string{}
    for i, net := range viper.Get(fmt.Sprintf("vms.%s.net", vmname)).([]any) {
        datamap[fmt.Sprintf("ipconfig%d", i)] = []string{""}
        if v, ok := net.(map[string]any)["network4"]; ok {
            ip, _ := getIP(userdata.IPIndex, v.(string))
            datamap[fmt.Sprintf("ipconfig%d", i)][0] += fmt.Sprintf("ip=%s,", ip)
        }
        if v, ok := net.(map[string]any)["gw4"]; ok {
            datamap[fmt.Sprintf("ipconfig%d", i)][0] += fmt.Sprintf("gw=%s,", v.(string))
        }
        if v, ok := net.(map[string]any)["network6"]; ok {
            ip, _ := getIP(userdata.IPIndex, v.(string))
            datamap[fmt.Sprintf("ipconfig%d", i)][0] += fmt.Sprintf("ip6=%s,", ip)
        }
        if v, ok := net.(map[string]any)["gw6"]; ok {
            datamap[fmt.Sprintf("ipconfig%d", i)][0] += fmt.Sprintf("gw6=%s,", v.(string))
        }
        l := len(datamap[fmt.Sprintf("ipconfig%d", i)][0])
        datamap[fmt.Sprintf("ipconfig%d", i)][0] = datamap[fmt.Sprintf("ipconfig%d", i)][0][:l-1]
    }
    if viper.GetBool(fmt.Sprintf("vms.%s.setuser", vmname)) {
        datamap["ciuser"] = []string{userdata.Username}
        if userdata.Password != "" {
            datamap["cipassword"] = []string{userdata.Password}
        }
        datamap["sshkeys"] = []string{strings.Replace(url.QueryEscape(user.GetSSHKey(c)), "+", "%20", -1)}
    }
    datamap["searchdomain"] = []string{viper.GetString(fmt.Sprintf("vms.%s.searchdomain", vmname))}
    datamap["nameserver"] = []string{viper.GetString(fmt.Sprintf("vms.%s.nameserver", vmname))}
    data := url.Values(datamap)
    req, _, err := pve.NewRequest(username, vmname, "PUT", useuri, strings.NewReader(data.Encode()))
    if err != nil {
        log.Panicln(err)
    }
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    //fmt.Println(req.URL)
    res, _ := pve.Start(req)
    defer res.Body.Close()
}

func getIP(index int64, cidr string) (ipAddress string, err error) {
	_, subnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return "", err
	}

	ip := big.NewInt(0).SetBytes(subnet.IP.To16()) // Ensure IP is in 16-byte representation
	ip.Add(ip, big.NewInt(int64(index)))

	ipBytes := ip.Bytes()
	// Ensure the IP bytes slice has the correct length for the type of IP address
	if len(subnet.IP) == net.IPv4len && len(ipBytes) > net.IPv4len {
		ipBytes = ipBytes[len(ipBytes)-net.IPv4len:]
	}
	if len(subnet.IP) == net.IPv6len && len(ipBytes) < net.IPv6len {
		ipBytes = append(make([]byte, net.IPv6len-len(ipBytes)), ipBytes...)
	}

	maskSize, _ := subnet.Mask.Size()

	return fmt.Sprintf("%s/%d", net.IP(ipBytes).String(), maskSize), nil
}
