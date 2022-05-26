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

app.config['SECRET_KEY'] = base64.b64decode('lO8irmEThaZ1TQSiouCxdGhejvUwgWd7mjJGVmp4'.encode())

nodedata = {}
with open('nodes.yaml', 'r') as f:
    nodedata = yaml.load(f)

uservmlist = {}
with open('uservmlist.yaml', 'r') as f:
    uservmlist = yaml.load(f)

def reloaddata():
    if os.path.isfile('.reload'):
        os.remove('.reload')
        global nodedata
        global uservmlist
        with open('nodes.yaml', 'r') as f:
            nodedata = yaml.load(f)
        with open('uservmlist.yaml', 'r') as f:
            uservmlist = yaml.load(f)


@app.route('/session',methods=['POST'])
def getsession():
    data = json.loads(request.data)
    session['username'] = data['username']
    return "success"


@app.route('/novnc/app/<path:path>',methods=['GET'])
def getdata(path):
    reloaddata()
    if session['username'] not in uservmlist:
        return Response("permission denied", 403, {})

    username=session['username']
    
    if 'admin' in uservmlist[username] and uservmlist[username]['admin']:
        if 'su' in session:
            username = session['su']
    if username not in uservmlist:
        return Response("permission denied", 403, {})

    r = requests.request('POST', "https://" + nodedata[uservmlist[username]['node']]['endpoint'] + "/api2/json/access/ticket", data='username=' + nodedata[uservmlist[username]['node']]['username'] + '&password=' + nodedata[uservmlist[username]['node']]['password'], verify=False)
    login = json.loads(r.content)
    r = requests.request('GET', "https://" + nodedata[uservmlist[username]['node']]['endpoint'] + "/novnc/app/" + path, headers={'cookie':'PVEAuthCookie=' + login['data']['ticket'], 'CSRFPreventionToken':login['data']['CSRFPreventionToken']}, verify=False)
    excluded_headers = ['content-encoding', 'content-length', 'transfer-encoding', 'connection']
    headers = [(name, value) for (name, value) in r.raw.headers.items() if name.lower() not in excluded_headers]
    headers = dict(headers)
    return Response(r.content.replace(nodedata[uservmlist[username]['node']]['username'].encode('utf-8'), b'').replace(login['data']['CSRFPreventionToken'].encode('utf-8'), b'').replace(login['data']['ticket'].encode('utf-8'), b''), r.status_code, headers)

@app.route('/vm',methods=['GET'])
def showvm():
    reloaddata()
    if session['username'] not in uservmlist:
        return Response("permission denied", 403, {})

    username=session['username']

    if 'admin' in uservmlist[username] and uservmlist[username]['admin']:
        if 'su' in request.args:
            if request.args['su'] not in uservmlist:
                return Response("permission denied", 403, {})
            session['su'] = request.args['su']
            username = request.args['su']
        else:
            del session['su']

    r = requests.request('POST', "https://" + nodedata[uservmlist[username]['node']]['endpoint'] + "/api2/json/access/ticket", data='username=' + nodedata[uservmlist[username]['node']]['username'] + '&password=' + nodedata[uservmlist[username]['node']]['password'], verify=False)
    login = json.loads(r.content)
    r = requests.request('GET', "https://" + nodedata[uservmlist[username]['node']]['endpoint'] + "/?console=kvm&vmid=" + uservmlist[username]['vmid'] + '&node=' + nodedata[uservmlist[username]['node']]['node'] + '&resize=scale&novnc=1', headers={'cookie':'PVEAuthCookie=' + login['data']['ticket'], 'CSRFPreventionToken':login['data']['CSRFPreventionToken']}, verify=False)
    excluded_headers = ['content-encoding', 'content-length', 'transfer-encoding', 'connection']
    headers = [(name, value) for (name, value) in r.raw.headers.items() if name.lower() not in excluded_headers]
    headers = dict(headers)
    return Response(r.content.replace(nodedata[uservmlist[username]['node']]['username'].encode('utf-8'), b'').replace(login['data']['CSRFPreventionToken'].encode('utf-8'), b'').replace(login['data']['ticket'].encode('utf-8'), b''), r.status_code, headers)

@app.route('/novnc/app.js',methods=['GET'])
def getapp():
    reloaddata()
    with open('app.' + request.args['ver'] + '.js', 'r') as f:
        data = f.read()
    return data

@app.route('/novnc/package.json',methods=['GET'])
def getpackage():
    reloaddata()
    if session['username'] not in uservmlist:
        return Response("permission denied", 403, {})

    username=session['username']
    
    if 'admin' in uservmlist[username] and uservmlist[username]['admin']:
        if 'su' in session:
            username = session['su']
    if username not in uservmlist:
        return Response("permission denied", 403, {})

    r = requests.request('GET', "https://" + nodedata[uservmlist[username]['node']]['endpoint'] + "/novnc/package.json", verify=False)
    excluded_headers = ['content-encoding', 'content-length', 'transfer-encoding', 'connection']
    headers = [(name, value) for (name, value) in r.raw.headers.items() if name.lower() not in excluded_headers]
    headers = dict(headers)
    return Response(r.content.replace(nodedata[uservmlist[username]['node']]['username'].encode('utf-8'), b''), r.status_code, headers)

@app.route('/vm/api/status/<cmd>',methods=['GET','POST'])
def getvmstatus(cmd):
    reloaddata()
    if session['username'] not in uservmlist:
        return Response("permission denied", 403, {})

    username=session['username']
    
    if 'admin' in uservmlist[username] and uservmlist[username]['admin']:
        if 'su' in session:
            username = session['su']
    if username not in uservmlist:
        return Response("permission denied", 403, {})

    r = requests.request('POST', "https://" + nodedata[uservmlist[username]['node']]['endpoint'] + "/api2/json/access/ticket", data='username=' + nodedata[uservmlist[username]['node']]['username'] + '&password=' + nodedata[uservmlist[username]['node']]['password'], verify=False)
    login = json.loads(r.content)
    r = requests.request(request.method, "https://" + nodedata[uservmlist[username]['node']]['endpoint'] + "/api2/json/nodes/" + nodedata[uservmlist[username]['node']]['node'] + "/qemu/" + uservmlist[username]['vmid'] + "/status/" + cmd, headers={'cookie':'PVEAuthCookie=' + login['data']['ticket'], 'CSRFPreventionToken':login['data']['CSRFPreventionToken']}, verify=False)
    excluded_headers = ['content-encoding', 'content-length', 'transfer-encoding', 'connection']
    headers = [(name, value) for (name, value) in r.raw.headers.items() if name.lower() not in excluded_headers]
    headers = dict(headers)
    return Response(r.content, r.status_code, headers)

#@app.route('/vm/api/<vmid>/vncwebsocket',methods=['GET'])
@app.route('/vm/api/vncproxy',methods=['POST'])
def vncconnect():
    reloaddata()
    if session['username'] not in uservmlist:
        return Response("permission denied", 403, {})

    username=session['username']
    
    if 'admin' in uservmlist[username] and uservmlist[username]['admin']:
        if 'su' in session:
            username = session['su']
    if username not in uservmlist:
        return Response("permission denied", 403, {})

    r = requests.request('POST', "https://" + nodedata[uservmlist[username]['node']]['endpoint'] + "/api2/json/access/ticket", data='username=' + nodedata[uservmlist[username]['node']]['username'] + '&password=' + nodedata[uservmlist[username]['node']]['password'], verify=False)
    login = json.loads(r.content)
    r = requests.request('POST', "https://" + nodedata[uservmlist[username]['node']]['endpoint'] + "/api2/json/nodes/" + nodedata[uservmlist[username]['node']]['node'] + "/qemu/" + uservmlist[username]['vmid'] + "/vncproxy", headers={'cookie':'PVEAuthCookie=' + login['data']['ticket'], 'CSRFPreventionToken':login['data']['CSRFPreventionToken']}, data='websocket=1', verify=False)
    return Response(json.dumps({'host':nodedata[uservmlist[username]['node']]['endpoint'],'nodes':nodedata[uservmlist[username]['node']]['node'],'vmid':uservmlist[username]['vmid'],'login':login,'vnc':json.loads(r.content)}), 200, {'Content-Type':'application/json;charset=UTF-8'})


if __name__ == "__main__":
#    socketio.run(app, host="0.0.0.0", port=101)
    app.run(host="127.0.0.1", port=4001)
