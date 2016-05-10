#!/usr/bin/env python

import json
import codecs
from subprocess import Popen, PIPE
from urllib.parse import urlparse, urlencode, parse_qs
from urllib.request import urlopen, Request

serveraddr = '127.0.0.1'

authorization_endpoint = 'http://'+serveraddr+'/oauth2/auth'
token_endpoint = 'http://'+serveraddr+'/oauth2/token'
userinfo_endpoint = 'http://'+serveraddr+'/api/me'
client_id = '1'
client_secret = 'admin'
redirect_uri = 'https://sso-test.com'
utf8reader = codecs.getreader('utf8')

url = authorization_endpoint + '?' + urlencode({
    'response_type': 'code',
    'redirect_uri': redirect_uri,
    'realm' : 'your-realms',
    'client_id': client_id,
    'scope': 'asdfasdfasdfasdf',
    'state': '0.847524793818593',
})

print ("url is : ",url)

username = "xuyn"
password = "1234"


p = Popen(['curl', '-v', url, '-d', 'login='+username+'&password='+password],
          stdout=PIPE, stderr=PIPE)
for line in p.stderr:
    line = line.decode('utf8')
    print(line.rstrip())
    if line.startswith('< Location: '):
        code_callback_url = line.strip()[len('< Location: '):]
        break
else:
    print(p.stdout.read())

authentication = parse_qs(urlparse(code_callback_url).query)
code = authentication['code'][0]
print("Code: ", code)

qs = urlencode({
    'code': code,
    'client_id': client_id,
    'client_secret': client_secret,
    'redirect_uri': redirect_uri,
    'grant_type': 'authorization_code',
})
print("> POST {}?{}".format(token_endpoint, qs.encode('utf8')))
f = utf8reader(urlopen(token_endpoint, qs.encode('utf8')))
accessinfo = json.load(f)
print(accessinfo)
access_token = accessinfo['access_token']
print("Access Token: ", access_token)

print("> GET {}".format(userinfo_endpoint))
req = Request(userinfo_endpoint, headers={
    'Authorization': 'Bearer {}'.format(access_token),
})
f = utf8reader(urlopen(req))
userinfo = json.load(f)
print(userinfo)
print("Username:", userinfo['name'])
print("Email:", userinfo['email'])
print("Groups:", ", ".join(userinfo['groups']))
