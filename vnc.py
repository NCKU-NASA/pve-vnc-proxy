import os
import json
import uuid
import base64
import zipfile
import io
import re
import requests
import yaml

from flask import Flask,request,redirect,Response,make_response,jsonify,render_template,session,send_file

app = Flask(__name__)

app.config['SECRET_KEY'] = os.urandom(24)

nodedata = {}
with open('nodes.yaml', 'r') as f:
    nodedata = yaml.load(f)

uservmlist = {}
with open('uservmlist.yaml', 'r') as f:
    uservmlist = yaml.load(f)

def reload_data():
    if os.path.isfile('.reload'):
        os.remove('.reload')
        global nodedata
        global uservmlist
        with open('nodes.yaml', 'r') as f:
            nodedata = yaml.load(f)
        with open('uservmlist.yaml', 'r') as f:
            uservmlist = yaml.load(f)

def getlogin(username):
    r = requests.request('POST', "https://" + nodedata[uservmlist[username]['node']]['endpoint'] + "/api2/json/access/ticket", data='username=' + nodedata[uservmlist[username]['node']]['username'] + '&password=' + nodedata[uservmlist[username]['node']]['password'], verify=False)
    return json.loads(r.content)

def sendrequest(username, login, method, path, data=None, headers=None):
    loginheader = None
    if login != None:
        loginheader = {'cookie':'PVEAuthCookie=' + login['data']['ticket'], 'CSRFPreventionToken':login['data']['CSRFPreventionToken']}
        if headers != None:
            loginheader = { **loginheader, **headers }
    else:
        if headers != None:
            loginheader = headers

    r = requests.request(method, "https://" + nodedata[uservmlist[username]['node']]['endpoint'] + path, headers=loginheader, data=data, verify=False)
    
    excluded_headers = ['content-encoding', 'content-length', 'transfer-encoding', 'connection']
    reqheaders = [(name, value) for (name, value) in r.raw.headers.items() if name.lower() not in excluded_headers]
    reqheaders = dict(reqheaders)
    content = r.content.replace(nodedata[uservmlist[username]['node']]['username'].encode('utf-8'), b'')
    if login != None:
        content = content.replace(login['data']['CSRFPreventionToken'].encode('utf-8'), b'').replace(login['data']['ticket'].encode('utf-8'), b'')
    response = Response(content, r.status_code, reqheaders)
    return {'request':r,'response':response}

def getusername(sufromsession):
    reload_data()
    if 'connect.sid' not in request.cookies:
        session.clear()

    if session.get('username') not in uservmlist:
        return Response("permission denied", 403, {})
    
    username=session.get('username')
    if 'admin' in uservmlist[username] and uservmlist[username]['admin']:
        if sufromsession:
            if 'su' in session:
                username = session.get('su')
        else:
            args = {}
            if request.method == 'GET':
                args = request.args
            elif request.method == 'POST':
                args = request.form
            if 'su' in args:
                if args['su'] not in uservmlist:
                    return Response("permission denied", 403, {})
                session['su'] = args['su']
                username = args['su']
            elif 'su' in session:
                del session['su']

    if username not in uservmlist:
        return Response("permission denied", 403, {})
    return username


@app.route('/session',methods=['POST'])
def getsession():
    data = json.loads(request.data)
    if 'username' not in data:
        return "fail"

    if 'username' in session and session['username'] == data['username']:
        return "success"

    session['username'] = data['username']
    return "success"

@app.route('/novnc/app/<path:path>',methods=['GET'])
def getdata(path):
    username = getusername(True)
    if type(username) == Response:
        return username

    login = getlogin(username)
    r = sendrequest(username, login,'GET',"/novnc/app/" + path)
    return r['response']

@app.route('/vm',methods=['GET'])
def showvm():
    username = getusername(False)
    if type(username) == Response:
        return username

    login = getlogin(username)
    r = sendrequest(username, login,'GET',"/?console=kvm&vmid=" + uservmlist[username]['vmid'] + '&node=' + nodedata[uservmlist[username]['node']]['node'] + '&resize=scale&novnc=1')
    return r['response']


@app.route('/novnc/app.js',methods=['GET'])
def getapp():
    username = getusername(True)
    if type(username) == Response:
        return username
    with open('app/' + nodedata[uservmlist[username]['node']]['app'], 'r') as f:
        data = f.read()
    return data

@app.route('/novnc/package.json',methods=['GET'])
def getpackage():
    username = getusername(True)
    if type(username) == Response:
        return username

    r = sendrequest(username, None,'GET', "/novnc/package.json")
    return r['response']

@app.route('/vm/api/status/<cmd>',methods=['GET','POST'])
def getvmstatus(cmd):
    username = getusername(False)
    if type(username) == Response:
        return username

    login = getlogin(username)
    r = sendrequest(username, login, request.method,"/api2/json/nodes/" + nodedata[uservmlist[username]['node']]['node'] + "/qemu/" + uservmlist[username]['vmid'] + "/status/" + cmd)
    return r['response']

@app.route('/vm/api/vncproxy',methods=['POST'])
def vncconnect():
    username = getusername(True)
    if type(username) == Response:
        return username

    login = getlogin(username)
    r = sendrequest(username, login, 'POST', "/api2/json/nodes/" + nodedata[uservmlist[username]['node']]['node'] + "/qemu/" + uservmlist[username]['vmid'] + "/vncproxy", data='websocket=1')
    return Response(json.dumps({'host':nodedata[uservmlist[username]['node']]['endpoint'],'nodes':nodedata[uservmlist[username]['node']]['node'],'vmid':uservmlist[username]['vmid'],'login':login,'vnc':json.loads(r['request'].content)}), 200, {'Content-Type':'application/json;charset=UTF-8'})


if __name__ == "__main__":
    app.run(host="127.0.0.1", port=4001)
