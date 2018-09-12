# SSO

SSO，Single Sign On系统，是一种身份验证和授权系统，使应用程序能够获得对HTTP服务上的用户帐户的有限访问。


##主要概念
   - 组：组是用户的集合。组可以拥有类似组的层次结构。一个父亲组可以有一些儿子组，但是一个儿子组只能有一个父组。组之间的关系可以是管理员或普通成员。如果用户是组A中的管理员，组A是组B的管理员和父组，我们可以说用户也是组B的管理员。
  
   - 资源：资源可以由用户定义。资源属于一个app。对于一个应用程序，可以将资源分配给app的角色。
  
   - 角色：角色是一组用户，属于一个客户端。角色组中的用户可以获得角色的资源。角色可以像组一样具有层次结构。一个父亲角色可以有一些儿子角色，但一个儿子角色只能有一个父亲角色。一个app至少有一个角色，即根角色。客户端中的所有其他角色都是root角色的子角色。父角色的用户可以访问其子角色的资源。只能为资源分配leaf角色。
  
   - 客户端：客户端是在SSO系统中注册的一个应用程序。在SSO中注册客户端时，客户的所有者可以获得密码和app id。 Secret和app id用于用户身份验证和授权。如果您想进一步了解它，可以阅读https://oauth.net/2/
  
   - 申请：用户可以通过提交一个申请来申请加入群组或者角色。应用程序系统将向这些组的管理员发送电子邮件，并让他们批准或拒绝该申请。

## install
```sh
   go get github.com/laincloud/sso
```

## usage
 ```sh
#r/local/bin/python2
import time
import requests
import json
from urlparse import urlparse, parse_qs
scope = 'write:role read:role write:resource read:resource'
payload = {'client_id':'clientId','response_type':'code','scope':scope,'redirect_uri':'redirectUri','state':'foobar'}
#clientId is the app id, which is generated in when the app is registered
#redirectUri is the the uri that set when the app is registered, which is used for redirect to your app from sso
usr_msg={'login':'username','password':'password'}
#username is your username of SSO and password is your password of SSO.
auth_url=ssohost+'/oauth2/auth'
#ssohost is the sso host domain
result=requests.post(auth_url,params=payload,data=usr_msg,allow_redirects=False)
code_callback_url=result.headers['location']
authentication=parse_qs(urlparse(code_callback_url).query)
code=authentication['code'][0]
auth_msg={'client_id':'clientId','grant_type':'authorization_code','client_secret':'clientSecret','code':code,'redirect_uri':'redirectUri'}
#clientSecret is the secret of the client, which is generated in when the app is registered.
result=requests.request("POST", ssohost + '/oauth2/token',headers=None,params=auth_msg)
accessinfo=result.json()
refresh_token=accessinfo['refresh_token']
auth_msg={'client_id':'clientId','grant_type':'refresh_token','client_secret':'clientSecret','refresh_token':refresh_token,'redirect_uri':'redirectUri'}
result = requests.request("GET",ssohost + '/oauth2/token',headers=None,params=auth_msg)
accessinfo=result.json()
token = 'Bearer '+accessinfo['access_token']
header = {'Authorization':token}
header2 = {'secret':'clientSecret'}
payload = {'app_id':'clientId','type':'raw'}
#clientId is the app id, which is generated in when the app is registered
createResource = {'name':'tester','description':'testing','data':'testing'}
updateResource = {'name':'tester2','description':'testing2','data':'testing2'}
createRole = {'app_id':clientId,'name':'test3','parent_id':roleId,'description':'testing3'}
#roleId is the id of father role of the role you are creating
updateRole = {'parent_id':roleId,'name':'test4','description':'test4'}
addMember = {'type':'normal'}
deleteResourceFromRole = {'action':'delete','resource_list':[id1,id2,id3]}
#id1,id2,id3 are the ids of resource that you want to delete 
addResourceToRole = {'action':'update','resource_list':[id4,id5,id6]}
#id4,id5,id6 are the ids of resource that you want to add
addMembersAccumulatively = {'Action':'add','RoleId':roleId,'members':[{'user':'name1','type':'normal'},{'user':'name2','type':'normal'}]}
#name1 and name2 are names of users that you want to add to the role
deleteResourceAccumulatively = [id7,id8]
#id7, id8 are the ids of resource you want to delete

print("testing add members accumulatively")
r = requests.post(ssohost + '/api/rolemembers',data=json.dumps(addMembersAccumulatively),headers=header)

print("testing create resource")
r = requests.post(ssohost + '/api/resources',params=payload,data=json.dumps(createResource),headers=header)

print("testing update resource")
r = requests.post(ssohost + '/api/resources/id9',params=payload,data=json.dumps(updateResource),headers=header)
#id9 is the id of resource that you want to update

print("testing delete resource accumulatively")
r = requests.post(ssohost + '/api/resourcesdelete',params=payload,headers=header,data=json.dumps(deleteResourceAccumulatively))

print("testing delete resource")
r = requests.delete(ssohost + '/api/resources/id10',params=payload,headers=header)
#id10 is the id of resource that you want to delete

print("testing get rosources of app")
r = requests.get(ssohost + '/api/resources',params=payload,headers=header)

print("testing create role")
r = requests.post(ssohost + '/api/roles',params=payload,data=json.dumps(createRole),headers=header)

print("testing update role")
r = requests.post('https://sso-ldapyifan.yxapp.xyz/api/roles/roleId',params=payload,headers=header,data=json.dumps(updateRole))
#roleId is the id of role you are updating

print("testing get role")
r = requests.get(ssohost + '/api/roles/roleId',params=payload,headers=header)
#roleId is the id of the role you want to get

print("testing get roles")
r = requests.get(ssohost + '/api/roles',params=payload,headers=header)

print("testing delete role")
r = requests.delete(ssohost + '/api/roles/roleId',params=payload,headers=header)
#roleId is the id of the role you want to delete

print("testing add member")
r = requests.put(ssohost + '/api/roles/roleId/members/name',params=payload,headers=header,data=json.dumps(addMember))
#roleId is the id of the role you want to add member
#name is the username of the user who you want to add

print("testing delete member")
r = requests.delete(ssohost + '/api/roles/roleId/members/name',params=payload,headers=header)
#roleId is the id of the role you want to delete member
#name is the username of the user who you want to delete

print("testing add resource to role")
r = requests.post(ssohost + '/api/roles/roleId/resources',params=payload,headers=header,data=json.dumps(addResourceToRole))
#roleId is the id of the role you want to add resouce
#note: resource can only be added to the leaf role

print("testing delete resource from role")
r = requests.post(ssohost + '/api/roles/roleId/resources',params=payload,headers=header,data=json.dumps(deleteResourceFromRole))
#roleId is the id of the role you want to delete resouce
```


