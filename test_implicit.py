#!/usr/bin/env python

import json
import codecs
from subprocess import Popen, PIPE
from urllib.parse import urlparse, urlencode, parse_qs
from urllib.request import urlopen, Request

serveraddr = 'localhost:8011'

authorization_endpoint = 'http://'+serveraddr+'/oauth2/auth'
token_endpoint = 'http://'+serveraddr+'/oauth2/token'
userinfo_endpoint = 'http://'+serveraddr+'/api/me'
client_id = '4'
client_secret = '-Ljt2Nt-XqOZlifx4_v0mg'
redirect_uri = 'http://localhost:8180'
utf8reader = codecs.getreader('utf8')

url = authorization_endpoint + '?' + urlencode({
    'response_type': 'id_token token', #可以只为 'token'
    'redirect_uri': redirect_uri,
    'client_id': client_id,
    'scope': 'openid',
    'state': '0.847524793818593',
})

print ("url is : ",url)

username = "cy"
password = "123456"


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

print(code_callback_url)
print(parse_qs(urlparse(code_callback_url).fragment))

accessinfo = parse_qs(urlparse(code_callback_url).fragment)

access_token = accessinfo['access_token'][0]
print("Access Token: ", access_token)

print("> GET {}".format(userinfo_endpoint))
req = Request(userinfo_endpoint, headers={
    'Authorization': 'Bearer {}'.format(access_token),
})
print(req)
f = utf8reader(urlopen(req))
userinfo = json.load(f)
print(userinfo)
print("Username:", userinfo['name'])
print("Email:", userinfo['email'])
print("Groups:", ", ".join(userinfo['groups']))
