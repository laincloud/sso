## sso权限管理相关接口文档

# 一、全局状态码说明
API接口全局请求基于http状态码判断返回结果

| 状态码 | 状态码说明     
| ----- | ----------- 
| 200   | 请求成功 
| 204   | 成功但无返回数据
| 400   | 请求参数错误 
| 401   | access token 失效
| 500   | 服务器端错误

# 二、接口返回格式说明
API接口返回数据统一使用JSON格式，返回格式为3种情况

1. 请求成功且返回数据(200)
    返回对应接口的JSON数据
2. 请求成功但无返回数据（204）
    返回空
3. 请求失败（除200，204以外全部code）
    返回字符串，描述发生的错误


# 三、接口

**Base Domain:**  `sso.yxapp.in（线上） sso.yxapp.xyz（测试）`
### 鉴权方式
 - **除登录之外所有接口** `Header Authorization字段传access token`  用于用户逻辑的鉴权

 - **部分接口** `Header secret字段传secret` 用于系统层面鉴权

1. 基础接口  
    [1.1 获取token](#1-1-)  
    [1.2 刷新access token](#1-2-)  
    [1.3 获取当前用户信息](#1-3-)  
  
2. App管理  
    [2.1 查询App列表](#2-1-)  
    [2.2 创建App](#2-2-)  
    [2.3 给App创建根角色](#2-3-)      
    [2.4 查询app基本信息](#2-4-)      
    [2.5 更新APP](#2-5-)      
    [2.6 查询App](#2-6-)  
    [2.7 查询App列表](#2-7-)

3. 资源管理  
    [3.1 查询资源](#3-1-)    
    [3.2 修改资源](#3-2-)  
    [3.3 删除资源](#3-3-)  
    [3.4 批量删除资源](#3-4-)     
    [3.5 批量新增资源](#3-5-)     
    [3.6 批量更新资源](#3-6-)

4. 角色管理  
    [4.1 查询角色树](#4-1-)  
    [4.2 新增子角色](#4-2-)  
    [4.3 修改角色信息](#4-3-)  
    [4.4 删除角色](#4-4-)

5. 用户角色管理  
    [5.1 查询某角色的用户列表](#5-1-)  
    [5.2 向某角色内添加或更新用户](#5-2-)  
    [5.3 向某角色内删除用户](#5-3-)  
    [5.4 向角色内批量添加或删除用户](#5-4-)

6. 角色和资源关联管理  
    [6.1 查询App下全部角色资源关联](#6-1-)    
    [6.2 批量管理角色权限关联关系](#6-2-)  

7. 工单系统     
    [7.1 发起申请](#7-1-)   
    [7.2 查看申请](#7-2-)   
    [7.3 审批申请](#7-3-) 
    
8. 用户管理     
    [8.1 查看用户](#8-1-)   
    [8.2 删除用户](#8-2-)   
    [8.3 批量查看](#8-3-)   
    [8.4 查看用户列表](#8-4-)     
    [8.5 查看自己](#8-5-) 

9. 组管理     
    [9.1 查看用户组](#9-1-)   
    [9.2 新增组](#9-2-)   
    [9.3 查看指定组](#9-3-)   
    [9.4 删除组](#9-4-)    
    [9.5 查看用户在某组成员类型](#9-5-)   
    [9.6 添加或者更新某组用户](#9-6-)    
    [9.7 删除某组内用户](#9-7-)   
    [9.8 添加组的附属关系](#9-8-)   
    [9.9 删除组的附属关系](#9-9-)   

## 1.基础接口
### 1.1 获取token

**接口地址及请求方式**  
`GET /oauth2/token?client_id=param&client_secret=param&redirect_uri=param&grant_type=param&oauthUrl=param&code=param`

**请求数据格式**  
`GET urlencode`

**请求参数（url）**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- | 
client_id|app的id|是|string|123|
client_secret|secret|是|string|BdksdfjsdJFudfk-s2
redirect_uri|重定向uri|是|string|http://xx.xx.xx
oauthUrl|oauthUrl|是|string|https://sso-ldap.yxapp.in/oauth2/auth|默认值
code|授权临时code|是|string|qHn0-lfIQ6e_VqxHMZCf4g|

**返回结果**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- | 
access_token|鉴权token|是|string|gHZycwzXRHivuNsTtxw2UrA|
expires_in|token有效时间|是|integer|1296000|
refresh_token|refresh token|是|integer|KuqB_2jfTss9WnDD_DtliQ|用于重新获取access token
scope|授权域|是|string|write:app read:app|access token的授权范围

**返回结果示例**  
```
{ 
    "access_token":"gHZycwzXRHivuNsTtxw2UrA",
    "expires_in":1296000,
    "refresh_token":"KuqB_2jfTss9WnDD_DtliQ",
    "scope":"write:app read:app read:user write:user write:group read:group write:role read:role write:resource read:resource",
    "token_type":"Bearer"
    
}
```

### 1.2 刷新access token

**接口地址及请求方式**  
`GET `

**请求数据格式**  


**请求参数**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- | 


**返回结果**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- | 

**返回结果示例**

```

```

### 1.3 获取当前用户信息

**接口地址及请求方式**  
`GET /api/me`

**请求数据格式**  


**请求参数**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- | 
无|

**返回结果**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- | 
res_code|返回状态|是|integer|0||
res_msg|返回信息|是|string|ok|
data|返回数据|是|object||

**返回data项结构**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- | 
email|公司邮箱|是|string|xxx@example.com||
fullname|中文名|是|string|张三||
name|邮箱前缀|是|string|xxx||
groups|用户所在的组|是|array|||

**返回结果示例**

```
{
  "res_code": 0,
  "res_msg": "ok",
  "data": {
    "email": "xxx@example.com",
    "fullname": "张三",
    "name": "xxx",
    "groups": [
      "group1",
      "group2",
    ]
  }
}

```


## 2. App管理 
### 2.1 查询App列表

注：查询用于权限控制的app列表，该类app有相应的root role

**接口地址及请求方式**  
`GET /api/app_roles`

**请求数据格式**
字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |

**请求参数**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- | 
无|

**返回结果**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- | 
id|app的id|是|integer|48|
fullname|app名称|是|string|
roles|角色|是|array||当前access token所对应用户在该app下的角色


**返回roles项结构**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- | 
id|角色id|是|integer|797|
name|角色名|是|string|pallas-transfer-sys|
type|角色类型|是|string|admin normal|当前用户在该角色下的类型
parent_id|父角色id|是|integer|123 |-1：当前角色是根角色

**返回结果示例**  

```
[
  {
    "id": 48,
    "fullname": "pallas-transfer",
    "roles": [
      {
        "name": "pallas-transfer-sys",
        "id": 797,
        "type": "admin",
        "parent_id": -1
      }
    ]
  }
]


```
### 2.2 创建App 
**接口地址及请求方式**  
  
```
POST /api/apps

body ：
{
    fullname:'test',
    redirect_uri:
}
```

**请求数据格式**  
`POST application/json`


**请求参数（body）**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- | 
fullname|app名称|是|string||
redirect_uri|重定向uri|是|string||

**返回结果**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- | 
id|app的id|是|integer|39||
fullname|app名称|是|string|test|
secret|secret|是|string||
redirect_uri|重定向uri|是|string|
admin_group|admin_group|是|object|

**返回admin_group项结构**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- | 
name|name|是|string||
fullname|fullname|是|string||


**返回结果示例**  

```
{
    "id": 39,
    "fullname": "test",
    "secret": "zmDAz38jq_iofdvLXPz3BQ",
    "redirect_uri": "http://test/222",
    "admin_group": {
        "name": ".app-39",
        "fullname": "App 39 Admin Group"
    }
}


```
### 2.3 给App创建根角色 
**接口地址及请求方式**  
`POST /api/app_roles `

**请求数据格式**  
`POST application/json`


**请求参数**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- | 
app_id|app的id|是|integer||
role_name|根角色名|是|string||

**返回结果**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- | 
Id|app的id|是|integer||
FullName|app的名称|是|string||
Secret|app的Secret|是|string||
RedirectUri|重定向uri|是|string||
AdminGroupId|AdminGroupId|是|string||
AdminRoleId|AdminRoleId|是|AdminRoleId||
Created|Created|是|Created||
Updated|Updated|是|Updated||

**返回结果示例**  

```
{
    "Id": 39,
    "FullName": "test",
    "Secret": "",
    "RedirectUri": "http://test/222",
    "AdminGroupId": 661,
    "AdminRoleId": 661,
    "Created": "2018-06-29 09:30:32",
    "Updated": "2018-06-29 09:38:00"
}


```


### 2.4 查询app基本信息
**接口地址及请求方式**
`GET /api/app_info `

**请求数据格式**



**请求参数**


**返回结果**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |
Id|app的id|是|integer||
FullName|app的名称|是|string||

**返回结果示例**

```
[
  {
    "id": 48,
    "fullname": "pallas-transfer",
  }
]


```

### 2.5 更新App
**接口地址及请求方式**

```
PUT /api/app/:id

body ：
{
    fullname:'test',
    redirect_uri:'http"//example'
}
```

**请求数据格式**
`PUT application/json`


**请求参数**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |
id|appid|是|int||
fullname|app名称|是|string||
redirect_uri|重定向uri|是|string||

**返回结果**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |
id|app的id|是|string|39||
fullname|app名称|是|string|test|
redirect_uri|重定向uri|是|string|


**返回结果示例**

```
{
    "id": 39,
    "fullname": "test",
    "redirect_uri": "http://test/222",
}


```


### 2.6 查询App
**接口地址及请求方式**

```
GET /api/app/:id

```

**请求数据格式**


**请求参数**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |
id|app的id|是|string|39||


**返回结果**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |
id|app的id|是|integer||
fullName|app的名称|是|string||
secret|app的Secret|是|string||
redirect_uri|重定向uri|是|string||

**返回结果示例**

```
{
    "id": 39,
    "fullname": "test",
    "secret": "eqwewqdcsdfsfdsf",
    "redirect_uri": "http://test/222",
}

```

### 2.7 查询App列表
**接口地址及请求方式**

```
GET /api/apps

```

**请求数据格式**

**请求参数**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |

**返回结果**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |
id|app的id|是|integer||
fullname|app的名称|是|string||
secret|app的Secret|是|string||
redirect_uri|重定向uri|是|string||
admin_group|管理组|是|struct||

**admin_group项参数**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |
name|组名|是|string||
fullname|组介绍|是|string||


**返回结果示例**

```
{
    "id": 39,
    "fullname": "test",
    "secret": "eqwewqdcsdfsfdsf",
    "redirect_uri": "http://test/222",
    "admin_group": {"name":"g1", "fullname":"test"}
}

```

## 3. 资源管理  
### 3.1 查询资源
**接口地址及请求方式**  
`GET /api/resources?app_id=123&type=raw`

note：查询App全部角色资源关联 `GET /api/resources?app_id=123&type=byrole`

**请求数据格式**  
`GET urlencode`

**请求参数**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- | 
app_id|app的id|是|integer||
type|类型|是|string|raw|固定传raw

**返回结果**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- | 
id|资源id|是|integer|1||
name|资源名|是|string||
description|资源描述|是|string||
app_id|app的id|是|integer||
data|资源内容|是|string||
owner|资源创建者|是|string||
created|创建时间|是|string||
updated|更新时间|是|string||




**返回结果示例**

```
[
  {
    "id": 1,
    "name": "测试权限1",
    "description": "权限1描述",
    "app_id": 124,
    "data": "test-1",
    "owner": "xxx",
    "created": "2018-05-22 18:31:59",
    "updated": "2018-05-22 18:31:59"
  },
  {
    "id": 2,
    "name": "测试权限2",
    "description": "权限2描述",
    "app_id": 124,
    "data": "test-2",
    "owner": "xxx",
    "created": "2018-05-23 13:54:10",
    "updated": "2018-05-23 13:54:10"
  }
]



```
### 3.2 新增资源
**接口地址及请求方式**  
`POST /api/resources?app_id=123 `

**请求数据格式**   
`POST application/json`



**请求参数**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- | 
name|资源名|是|string||
description|资源描述|是|string||
data|资源内容|是|string||

**返回结果**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- | 
id|资源id|是|integer|1||
name|资源名|是|string||
description|资源描述|是|string||
app_id|app的id|是|integer||
data|资源内容|是|string||
owner|资源创建者|是|string||
created|创建时间|是|string||
updated|更新时间|是|string||




**返回结果示例**

```
{
  "id": 14,
  "name": "测试权限3",
  "description": "权限描述",
  "app_id": 124,
  "data": "ddd",
  "owner": "xxx",
  "created": "2018-06-29 12:48:28",
  "updated": "2018-06-29 12:48:28"
}


```
### 3.3 修改资源
**接口地址及请求方式**  
`POST /api/resources/:id `

**请求数据格式**   
`POST application/json`


**请求参数**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- | 
name|资源名|是|string||
description|资源描述|是|string||
data|资源内容|是|string||

**返回结果**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- | 
id|资源id|是|integer|1||
name|资源名|是|string||
description|资源描述|是|string||
app_id|app的id|是|integer||
data|资源内容|是|string||
owner|资源创建者|是|string||
created|创建时间|是|string||
updated|更新时间|是|string||


**返回结果示例**

```
{
  "id": 14,
  "name": "测试权限3",
  "description": "权限描述",
  "app_id": 124,
  "data": "ddd",
  "owner": "xxx",
  "created": "2018-06-29 12:48:28",
  "updated": "2018-06-29 12:48:28"
}
```
### 3.4 删除资源
**接口地址及请求方式**  
`DELETE /api/resources/:id` 

**请求数据格式** 



**请求参数**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- | 
无|

**返回结果**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- | 
无




**返回结果示例**

`http code 204 no content`

### 3.5 批量删除资源
**接口地址及请求方式**  

`POST /api/resources?app_id=102&action=delete`

**请求数据格式** 

`POST application/json`


**请求参数**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- | 
resources|资源id数组|是|array|[{id:1},{id:2},{id:3}]||
app_id|删除哪个App资源下资源|是|string|10|url参数
action|操作|是|string|||

**返回结果**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- | 
无|


**返回结果示例**
`http code 204 no content`


### 3.6 批量新增资源
**接口地址及请求方式**  

`POST /api/resources?app_id=102&action=add`

**请求数据格式** 

`POST application/json`


**请求参数**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- | 
resources|资源id数组|是|array| | |
app_id|新增哪个App资源下资源|是|string|10|url参数
action|操作|是|string|||

**resources项参数**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- | 
name|资源名|是|string||
description|资源描述|是|string||
data|资源内容|是|string||

**返回结果**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- | 
无|


**返回结果示例**
`http code 200 status ok`


### 3.7 批量更新资源
**接口地址及请求方式**  

`POST /api/resources?app_id=102&action=update`

**请求数据格式** 

`POST application/json`


**请求参数**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- | 
resources|资源id数组|是|array| | |
app_id|更新哪个App资源下资源|是|string|10|url参数
action|操作|是|string|||

**resources项参数**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- | 
id|资源id|是|int|10||
name|资源名|是|string||
description|资源描述|是|string||
data|资源内容|是|string||

**返回结果**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- | 
无|


**返回结果示例**
`http code 200 status ok`


## 4. 角色管理  
### 4.1 查询角色树  
**接口地址及请求方式**  

`GET api/roles?app_id=134&all=true`

**请求数据格式**  

`GET urlencode`


**请求参数**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- | 
app_id|app的id|是|string||
all|是否要查看所有roles|否|string||
note:如果all=true，返回app下所有的roles，不返回role成员, 否则返回用户在该app下拥有的roles和role的members
**返回结果**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- | 
id|角色id|是|integer|797|
name|角色名|是|string|pallas-transfer-sys|
type|角色类型|是|string|admin normal|当前用户在该角色下的类型
parent_id|父角色id|是|integer|123 |-1：当前角色是根角色
app_id|app的id|是|integer||
created|创建时间|是|string||
updated|更新时间|是|string||
members|拥有该角色的用户|是|array||

**返回members项结构**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- | 
user|邮箱前缀|是|string||
type|用户在该角色下的类型|是|string|admin normal


**返回结果示例**

```
[
  {
    "id": 1768,
    "name": "xxx",
    "description": "xxx",
    "parent_id": -1,
    "app_id": 124,
    "created": "2018-05-04 20:10:43",
    "updated": "2018-05-04 20:10:43",
    "type": "admin",
    "members": [
      {
        "user": "xxx",
        "type": "admin"
      },
      {
        "user": "xxx",
        "type": "admin"
      },
      {
        "user": "xxx",
        "type": "admin"
      }
    ]
  },
  {
    "id": 2071,
    "name": "zw",
    "description": "xxx",
    "parent_id": 1768,
    "app_id": 124,
    "created": "2018-05-23 13:54:59",
    "updated": "2018-05-23 13:54:59",
    "type": "admin",
    "members": []
  },
  {
    "id": 2072,
    "name": "jhjj",
    "description": "xxx",
    "parent_id": 1768,
    "app_id": 124,
    "created": "2018-05-23 13:55:52",
    "updated": "2018-05-23 13:55:52",
    "type": "admin",
    "members": [
      {
        "user": "xxx",
        "type": "normal"
      }
    ]
  }
]



```
### 4.2 新增子角色  
**接口地址及请求方式**  

`POST /api/roles `

**请求数据格式**  

`POST application/json`


**请求参数**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- | 
qpp_id|app的id|是|integer||
description|角色描述|是|string||
name|角色名|是|string||
parent_id|父角色id|是|integer||


**返回结果**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- | 
id|角色id|是|integer|797|
name|角色名|是|string|pallas-transfer-sys|
type|角色类型|是|string|admin normal|当前用户在该角色下的类型
parent_id|父角色id|是|integer|123 |-1：当前角色是根角色
app_id|app的id|是|integer||
created|创建时间|是|string||
updated|更新时间|是|string||



**返回结果示例**

```
{
  "id": 2165,
  "name": "test",
  "description": "xxx",
  "parent_id": 2071,
  "app_id": 124,
  "created": "2018-06-29 13:08:29",
  "updated": "2018-06-29 13:08:29"
}


```
### 4.3 修改角色信息
**接口地址及请求方式**  

`POST api/roles/:id `

**请求数据格式**  

`POST application/json`


**请求参数**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- | 
description|角色描述|是|string||
name|角色名|是|string||
parent_id|父角色id|是|integer||

**返回结果**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- | 
id|角色id|是|integer|797|
name|角色名|是|string|pallas-transfer-sys|
type|角色类型|是|string|admin normal|当前用户在该角色下的类型
parent_id|父角色id|是|integer|123 |-1：当前角色是根角色
app_id|app的id|是|integer||
created|创建时间|是|string||
updated|更新时间|是|string||



**返回结果示例**

```
{
  "id": 2165,
  "name": "test",
  "description": "xxx",
  "parent_id": 2071,
  "app_id": 124,
  "created": "2018-06-29 13:08:29",
  "updated": "2018-06-29 13:08:29"
}


```
### 4.4 删除角色
**接口地址及请求方式**  

`DELETE /api/roles/:id `

**请求数据格式** 



**请求参数**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- | 
无|

**返回结果**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- | 
无|




**返回结果示例**

`http code 204 no content`

## 5. 用户角色管理  
### 5.1 查询某角色的用户列表（未使用）  
### 5.2 向某角色内添加或修改用户
**接口地址及请求方式**  
`PUT /api/roles/:role_id/members/:username`

**请求数据格式**  
`PUT `


**请求参数**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- | 
无|

**返回结果**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- | 
无|


**返回结果示例**

```
"member added"


```
### 5.3 向某角色内删除用户
**接口地址及请求方式**  

`DELETE  /api/roles/:role_id/members/:username`

**请求数据格式** 



**请求参数**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- | 
无|

**返回结果**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- | 
无|


**返回结果示例**

`http code 204`

### 5.4 向角色内批量添加或删除用户
**接口地址及请求方式**  

`POST api/rolemembers `

**请求数据格式**  

`POST application/json`


**请求参数**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- | 
roleId|角色id|是|integer||
action|操作|是|string|add delete
members|用户|是|array||
memebers:user|邮箱前缀|是|string||
memebers:type|用户在该角色下类型|是|string|


**返回结果**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- | 
res_code|返回状态|是|integer|0||
res_msg|返回信息|是|string|ok|
data|返回数据|是|object||

**返回data项结构**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- | 
无|

**返回结果示例**

```
"members added"
    or
"members deleted"


```

## 6. 角色和资源关联管理  
### 6.1 查询App下全部角色资源关联 
**接口地址及请求方式**  
`GET /api/resources?app_id=123&type=byrole`


注：3.1 查询App全部资源  `GET /api/resources?app_id=123&type=raw`

**请求数据格式** 



**请求参数**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- | 
app_id|app的id|是|integer||
type|类型|是|string|byrole|固定传byrole

**返回结果**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- | 
role_id|角色id|是|integer|1||
resources|关联的资源|是|array||


**返回resources项结构**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- | 
id|资源id|是|integer|1||
name|资源名|是|string||
description|资源描述|是|string||
app_id|app的id|是|integer||
data|资源内容|是|string||
owner|资源创建者|是|string||
created|创建时间|是|string||
updated|更新时间|是|string||

**返回结果示例**

```
[
  {
    "role_id": 2071,
    "resources": [
      {
        "id": 15,
        "name": "权限1",
        "description": "22",
        "app_id": 124,
        "data": "333",
        "owner": "xxx",
        "created": "2018-06-29 13:36:01",
        "updated": "2018-06-29 13:36:01"
      }
    ]
  }
]


```
### 6.2 批量管理角色和资源关联关系
**接口地址及请求方式**  
`POST /api/roles/:role_id/resources `

**请求数据格式** 



**请求参数**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- | 
action|操作|是|string|delete update add||
resource_list|资源id数组|array|[1,2,3]|

**返回结果**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- | 
无|




**返回结果示例**

`http code 204 no content`




## 7.工单接口

### 7.1 发起申请

**接口地址及请求方式**
`POST /api/applications`

**请求数据格式**
`POST application/json`

**请求参数**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |
target_type|申请的类型|是|string|group||
target|申请目标|是|[]struct|[{id:1,role:admin},{name:2,role:normal}]||
reason|申请理由|是|string||


**target项结构**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |
id| 申请项目id| 是 | int | 1 |  |
role| 申请项目职位| 是 | string | admin |  |


**返回结果**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |
id |申请id | 是 | integer | 3 |  |
applicant_email|申请人邮箱|是|string|sanzhang10@example.com||
target_type|申请的类型|是|string|group||
target|申请目标|是|array|{name:group1,role:admin}||
reason|申请理由|是|string||
status|申请状态|是|string||
commit_email|经办人邮箱|是|string||
created|创建时间|是|string||
updated|更新时间|是|string||

note:返回的status分为四类：initialled, approved, rejected, existed
创建一个新的申请，状态为initialled
处理一个申请，申请通过为approved，不通过为rejected
如果创建一个新的申请时，该申请已存在，状态为existed

**target项结构**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |
id| 申请项目id| 是 | int | 1 |  |
role| 申请项目职位| 是 | string | admin |  |
name|申请项目名称|是|string|||
app_name|申请role的时候用到，返回role所属的app name|否|string||

**返回结果示例**
```
{
 [
  {
    “id”: 3;
    “applicant_email”: “sanzhang10@example.com”,
    “target_type”:”group”,
    “target”:
		  {
		     “id”: 1,
             “role”:”admin”,
             "name": "group1"
           }
    “reason”:”理由”,
    “status”:”initialled”
    “commit_email”:["NULL"]
    “created”:"2018-06-29 12:48:28",
    "updated": "2018-06-29 12:48:28"
  }
 ]
}
```



### 7.2 查看申请
**接口地址及请求方式**
`GET api/applications?from=0&to=100&applicant_email=sanzhang3@example.com&status=approved&commit_email=exampleadmin@example.com`
**请求数据格式**
`GET urlencode`

**请求参数**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |
from|起始条数|否|integer|0|从最近的第几条
to|结束条数|否|integer|100|到最近的第几条
applicant_email|申请人邮箱|否|string|sanzhang3@example.com||
status|申请状态|否|string|approved||
commit_email|审批人邮箱|否|string|exampleadmin@example.com||

note：
如果from，to为空,默认发送最新的50条。
如果是申请人使用：applicant_email不填或者填自己的邮箱，返回申请人的申请（lain组必须填自己邮箱）
如果是审批人使用：填写commit_email,返回该审批人所要审批的申请
如果是lain组成员使用：填写applicant_email，返回该邮箱对应的申请，不填的话返回所有人的申请

**返回结果**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |
total|总条数|是|int||
applications|申请|是|array||

**applications项结构**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |
applicant_email|申请人邮箱|是|string|sanzhang10@example.com||
target_type|申请的类型|是|string|group||
target|申请目标|是|array|[{name:group1,role:admin}]||
reason|申请理由|是|string||
status|申请状态|是|string||
commit_emails|经办人邮箱|是|array(string)||
created|创建时间|是|string||
updated|更新时间|是|string||

**target项结构**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |
id| 申请项目id| 是 | int | 1 |  |
role| 申请项目职位| 是 | string | admin |  |
name|申请项目名称|是|string|||
app_name|申请role的时候用到，返回role所属的app name|否|string||


**返回结果示例**
```
{
"applications":[
    {
      “id”: 3;
      “applicant_email”: “sanzhang10@example.com”,
      “target_type”:”role”,
      “target”:[
                {
                “name”:”role1”,
                “role”: ”admin”,
                "id": 1,
                "app_name": "app1"
              }
                ]
      “reason”:”理由”,
      “status”:”initialled”
      “commit_emails”:["NULL"]
      “created”:"2018-06-29 12:48:28",
      "updated": "2018-06-29 12:48:28"
    }
  ]
"total":20
}
```

### 7.3 审批申请
**接口地址及请求方式**
`POST /api/applications/:application_id?action=approve`


note: action分为三类，approve，reject，recall
approve是审批人通过申请
reject是审批人拒绝申请
recall是申请人撤回申请，如果撤回成功则返回http code 204


**请求数据格式**
`POST urlencode`

**请求参数**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |
application_id|application number|是|integer|8||
action|处理审批|是|string|approve||

**返回结果**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |
applicant_email|申请人邮箱|是|string|sanzhang10@example.com||
target_type|申请的类型|是|string|group||
target|申请目标|是|array|[{name:group1,role:admin}]||
reason|申请理由|是|string||
status|申请状态|是|string||
commit_emails|经办人邮箱|是|array(string)||
created|创建时间|是|string||
updated|更新时间|是|string||


**target项结构**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |
id| 申请项目id| 是 | int | 1 |  |
role| 申请项目职位| 是 | string | admin |  |
name|申请项目名称|是|string|||
app_name|申请role的时候用到，返回role所属的app name|否|string||



**返回结果示例**
```
  {
    “id”: 3;
    “applicant_email”: “sanzhang10@example.com”,
    “target_type”:”role”,
    “target”:[
                {
                 “name”:”role1”,
                 “role”: ”admin”,
                 "id": 1,
                 "app_name": "app1"
                }
             ]
    “reason”:”理由”,
    “status”:”approved”
    “commit_emails”:["exampleadmin"]
    “created”:"2018-06-29 12:48:28",
    "updated": "2018-06-29 12:48:28"
  }
```

## 8.用户管理


### 8.1 查看用户
**接口地址及请求方式**
`GET /api/users/:username?database=true`


**请求数据格式**
`GET urlencode`

**请求参数**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |
username|查询用户名|是|string|||
database|是否只查数据库管理的组|否|string|true||

**返回结果**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |
user|用户档案|是|struct|||
groups|用户所在组|是|array|||


**user项结构**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |
name|用户名| 是 | string |  |  |
fullname|用户全名|是|string|||
mobile| 手机号| 否| string | |  |
email|邮箱|否|string|||
**返回结果示例**
```
{
    "user":{
        "name":"sanzhang
        "fullname:"张三"
        "mobile":"1xxxxxxxxxx"
        "email":"example@xxx.com"
        }
    "groups": [
        "group1"
        "group2"
        ]
}
```

### 8.2 删除用户
**接口地址及请求方式**
`DELETE /api/users/:username`


**请求数据格式**
`DELETE urlencode`

**请求参数**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |
username|查询用户名|是|string|||


**返回结果**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |



**返回结果示例**

`http code 204 no content`


### 8.3 批量查看
**接口地址及请求方式**
`GET api/batch-users`


**请求数据格式**
`GET urlencode`

**请求参数**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |
group|是否查询组|否|string|true||
name|查询用户名，用,隔开|是|string|sanzhang,sili,wuwang||

**返回结果**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |
profiles|用户们的档案|否|array|||
detailedProfiles|用户们的详细档案|否|array|||



**profiles项结构**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |
name|用户名| 是 | string |  |  |
fullname|用户全名|是|string|||
mobile| 手机号| 否| string | |  |
email|邮箱|否|string|||

**detailedProfiles项结构**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |
user|用户档案|是|array|||
groups|用户所在组|是|array|||


**user项结构**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |
name|用户名| 是 | string |  |  |
fullname|用户全名|是|string|||
mobile| 手机号| 否| string | |  |
email|邮箱|否|string|||
**返回结果示例**
```
{
    [ 
        {
            "name":"sanzhang
            "fullname:"张三"
            "mobile":"1xxxxxxxxxx"
            "email":"example@xxx.com"
            "groups": [
                "group1"
                "group2"
            ]
        }
    ]
}
```



### 8.4 查看用户列表
**接口地址及请求方式**
`GET api/users`


**请求数据格式**
`GET urlencode`

**请求参数**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |


**返回结果**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |
user|用户档案|是|struct|||
groups|用户所在组|是|array|||


**user项结构**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |
name|用户名| 是 | string |  |  |
fullname|用户全名|是|string|||
mobile| 手机号| 否| string | |  |
email|邮箱|否|string|||

**返回结果示例**
```
{
    [
        {
            "user":{
            "name":"sanzhang
            "fullname:"张三"
            "mobile":"1xxxxxxxxxx"
            "email":"example@xxx.com"
            }
            "groups": [
                "group1"
                "group2"
            ]
        }
    ]
}
```

### 8.5 查看自己
**接口地址及请求方式**
`GET api/me`


**请求数据格式**
`GET urlencode`

**请求参数**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |


**返回结果**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |
user|用户档案|是|struct|||
groups|用户所在组|是|array|||


**user项结构**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |
name|用户名| 是 | string |  |  |
fullname|用户全名|是|string|||
mobile| 手机号| 是| string | |  |
email|邮箱|是|string|||

**返回结果示例**
```
{
        {
            "user":{
            "name":"sanzhang
            "fullname:"张三"
            "mobile":"1xxxxxxxxxx"
            "email":"example@xxx.com"
            }
            "groups": [
                "group1"
                "group2"
            ]
        }
}
```


### 9.1 查看用户组
**接口地址及请求方式**
`GET api/groups`


**请求数据格式**
`GET urlencode`

**请求参数**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |


**返回结果**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |
name|组名|是|string||
fullname|组介绍|是|string||
role|成员类型|是|string|admin||

**返回结果示例**
```
{
    [
        {
            "name": "g1"
            "fullname:"test"
            "role":"admin"
        }
    ]
}
```

### 9.2 新增组
**接口地址及请求方式**
`POST api/groups`


**请求数据格式**
`POST application/json`

**请求参数**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |
group|新组|是|struct||
rules|规则|是|string||
backend|后端类型|是|integer||

**group项参数**
字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |
name|组名|是|string||
fullname|组介绍|是|string||


**返回结果**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |
name|组名|是|string||
fullname|组介绍|是|string||


**返回结果示例**
```
{
        {
            "name": "g1"
            "fullname:"test"
        }
}
```


### 9.3 查看指定组
**接口地址及请求方式**
`GET api/groups/:groupname`


**请求数据格式**
`GET urlencode`

**请求参数**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |
groupname|组名|是|string||

**返回结果**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |
name|组名|是|string||
fullname|组介绍|是|string||
members|组内用户|是|array||
group_members|子组列表|是|array||

**members项参数**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |
name|用户名|是|string||
role|成员类型|是|string||admin|


**group_members项参数**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |
name|子组名|是|string||
fullname|组介绍|是|string||
role|父组对子组的关系|是|string|admin||

**返回结果示例**
```
{
        {
            "name": "g1"
            "fullname:"test"
            "members": [
                         {
                            "name":"sanzhang"
                            "role":"admin"
                         }
                       ]
            "group_members": [
                                {
                                    "name":"g2"
                                    "fullname:"test"
                                    "role":"admin"
                                }
                             ]
        }
}
```


### 9.4 删除组
**接口地址及请求方式**
`DELETE api/groups/:groupname`


**请求数据格式**
`DELETE urlencode`

**请求参数**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |
groupname|组名|是|string||

**返回结果**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |


**返回结果示例**

`http code 204 no content`

### 9.5 查看用户在某组成员类型
**接口地址及请求方式**
`GET api/groups/:groupname/members/:username`


**请求数据格式**
`GET urlencode`

**请求参数**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |
groupname|组名|是|string||
username|成员名|是|string||

**返回结果**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |
role|成员类型|是|string|admin||

**返回结果示例**

```
{
   "role": "admin"
}
```

### 9.6 添加或者更新某组用户
**接口地址及请求方式**
`PUT api/groups/:groupname/members/:username`


**请求数据格式**
`PUT application/json`

**请求参数**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |
groupname|组名|是|string||
username|成员名|是|string||
role|成员类型|是|string||

**返回结果**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |

**返回结果示例**

### 9.7 删除某组内用户
**接口地址及请求方式**
`DELETE api/groups/:groupname/members/:username`


**请求数据格式**
`DELETE urlencode`

**请求参数**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |
groupname|组名|是|string||
username|成员名|是|string||

**返回结果**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |

**返回结果示例**

`http code 204 no content`


### 9.8 添加组的附属关系
**接口地址及请求方式**
`PUT api/groups/:groupname/group-members/:sonname`


**请求数据格式**
`PUT application/json`

**请求参数**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |
groupname|组名|是|string||
username|成员名|是|string||
role|子组成员类型|是|string||

**返回结果**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |

**返回结果示例**


### 9.9 删除组的附属关系
**接口地址及请求方式**
`DELETE api/groups/:groupname/group-members/:sonname`


**请求数据格式**
`DELETE urlencode`

**请求参数**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |
groupname|组名|是|string||
username|成员名|是|string||


**返回结果**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |

**返回结果示例**
`http code 204 no content`

