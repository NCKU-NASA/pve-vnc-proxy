import socket
import ssl
import requests
import json
import threading
import time
import urllib
from _thread import *

usessl = False
port = 4000
apiendpoint = ('127.0.0.1',4001)


server = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
server.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)

if usessl:
    server = ssl.wrap_socket(server, server_side=True, keyfile="path/to/keyfile", certfile="path/to/certfile")

apikeylist = {}

def startwebsuck(conn, pve):
    data = b''
    sessionon = True
    while sessionon:
        try:
            pve.settimeout(0.001)
            data = pve.recv(8192)
            if len(data) == 0:
                sessionon = False
            else:
#                print(data.decode('utf-8'))
                conn.settimeout(None)
                conn.send(data)
                data = b''
        except socket.timeout:
            data = b''
        try:
            conn.settimeout(0.001)
            data = conn.recv(8192)
            if len(data) == 0:
                sessionon = False
            else:
#                print(data.decode('utf-8'))
                pve.settimeout(None)
                pve.send(data)
                data = b''
        except socket.timeout:
            data = b''

    conn.close()
    pve.close()


def start():
    server.bind(('', port))
    server.listen(0)

    while True:
        try:
            conn, addr = server.accept()
            data = conn.recv(8192)
            
            if "Upgrade: websocket" in data.decode('utf-8'):
#            print(data.decode('utf-8'))
                alldata = data.decode('utf-8').splitlines()
                headers = {}
                for i in range(len(alldata)):
                    if i == 0:
                        headers['head'] = alldata[i].split()
                    elif ':' in alldata[i]:
                        line = alldata[i].split(':', 1)
                        headers[line[0].strip()] = line[1].strip()

                sessionon = True
                apikey=""
                urlparam = dict(urllib.parse.parse_qsl(urllib.parse.urlsplit(headers['head'][1]).query))
#            r = requests.request('GET', "http://" + apiendpoint[0] + ":" + str(apiendpoint[1]) + headers['head'][1])
#            apikey=json.loads(r.content)
                apikey=apikeylist[urlparam['vncticket']]
                del apikeylist[urlparam['vncticket']]
#                headers['head'][1] = headers['head'][1].replace('api','api2/json/nodes/' + apikey['nodes'] + '/qemu')
#            headers['head'][1] = '/api2/json/nodes/' + apikey['nodes'] + '/qemu/' + apikey['vmid'] + '/vncwebsocket?port=' + apikey['vnc']['data']['port'] + '&vncticket=' + apikey['vnc']['data']['ticket']
                headers['head'][1] = '/api2/json/nodes/' + apikey['nodes'] + '/qemu/' + apikey['vmid'] + '/vncwebsocket?port=' + urllib.parse.quote(apikey['vnc']['data']['port']) + '&vncticket=' + urllib.parse.quote(apikey['vnc']['data']['ticket'])
                headers['Origin'] = 'https://' + apikey['host']
                headers['Host'] = apikey['host']
#            headers['Cookie'] = 'PVEAuthCookie=' + urllib.parse.quote(apikey['login']['data']['ticket'])
                headers['Cookie'] = 'PVEAuthCookie=' + apikey['login']['data']['ticket']
                headers['CSRFPreventionToken'] = apikey['login']['data']['CSRFPreventionToken']
                
                data = ""
                for a in headers['head']:
                    data += a + " "
                data = data.strip() + '\r\n'
                del headers['head']
                for a in headers.items():
                    data += a[0] + ": " + a[1] + '\r\n'
                data += '\r\n'

#            data = "GET / HTTP/1.1\r\nHost: pve.ccns.io\r\nuser-agent: curl/7.74.0\r\naccept: */*\r\n\r\n"

                #print(apikey['vnc']['data']['ticket'])

                pve = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
                pve.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
                pve = ssl.wrap_socket(pve, ca_certs=None)
                if ':' in apikey['host']:
                    pve.connect((apikey['host'].rsplit(":", 1)[0], int(apikey['host'].rsplit(":", 1)[1])))
                else:
                    pve.connect((apikey['host'], 443))
                pve.settimeout(None)
                pve.send(data.encode('utf-8'))
                start_new_thread(startwebsuck, (conn, pve))
            else:
                api = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
                api.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
                api.connect(apiendpoint)
                api.settimeout(None)
                api.send(data)
                print(data.decode('utf-8').splitlines()[0])
                vncproxy=False
                if 'vncproxy' in data.decode('utf-8'):
                    vncproxy = True
                #print(data)
                data = b''
                sessionon = True
                gothead = False
                headers = {}
                allrecvdata = b''
                while sessionon:
                    try:
                        api.settimeout(0.001)
                        data = api.recv(8192)
                        if len(data) == 0:
                            if len(allrecvdata) > 0:
                                conn.settimeout(None)
                                conn.send(allrecvdata)
                            sessionon = False
                        else:
                            if not gothead:
                                allrecvdata += data
                                if b'\r\n\r\n' in allrecvdata:
                                    head=allrecvdata.split(b'\r\n\r\n',1)[0].decode('utf-8').splitlines()
                                    for i in range(len(head)):
                                        if i == 0:
                                            headers['head'] = head[i].split()
                                        elif ':' in head[i]:
                                            line = head[i].split(':', 1)
                                            headers[line[0].strip()] = line[1].strip()

                                    data = allrecvdata.split(b'\r\n\r\n',1)[1]
                                    allrecvdata = b''
                                    gothead = True
                                else:
                                    data = b''

                            if len(data) > 0:
                                if vncproxy:
                                    ticketdata = json.loads(data)
     #                               print(ticketdata)
                                    newticketdata = {'data':{'port':ticketdata['vnc']['data']['port'],'ticket':ticketdata['vnc']['data']['ticket']}}
                                    data = json.dumps(newticketdata).encode('utf-8')
                                    apikeylist[ticketdata['vnc']['data']['ticket']] = ticketdata


                                if len(headers) > 0:
                                    headers['Server'] = 'nginx'
                                    if vncproxy:
                                        headers['Content-Length'] = str(len(data))
                                    head = ""
                                    for a in headers['head']:
                                        head += a + " "
                                    head = head.strip() + '\r\n'
                                    del headers['head']
                                    for a in headers.items():
                                        head += a[0] + ": " + a[1] + '\r\n'
                                    head += '\r\n'
                                    conn.settimeout(None)
                                    conn.send(head.encode('utf-8'))
                                    headers = {}

                                conn.settimeout(None)
                                conn.send(data)
                                data = b''
                    except socket.timeout:
                        data = b''
                    try:
                        conn.settimeout(0.001)
                        data = conn.recv(8192)
                        if len(data) == 0:
                            sessionon = False
                        else:
#                        print(data.decode('utf-8'))
                            api.settimeout(None)
                            api.send(data)
                            data = b''
                    except socket.timeout:
                        data = b''

#            time.sleep(1)
                conn.close()
                api.close()
        except KeyboardInterrupt:
            exit()
        except:
            pass

if __name__ == "__main__":
    #thread_max = threading.BoundedSemaphore(1000)
    start()
