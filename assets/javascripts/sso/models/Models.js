import Cookie from './Cookie';

export let User = {
  
  resetPassword(username, callback) {
    Fetch.text(`/api/users/${username}/request-reset-password`, 'POST', null, null, (code, txt) => {
      callback && callback(code === 202, code === 202 ? `Succeed! ${txt}` : `Failed! ${txt}`);
    }, (msg) => {
      callback && callback(false, msg); 
    });
  },
  
  resetPasswordByEmail(formData, callback) {
    Fetch.text(`/api/request-reset-password-by-email`, 'POST', null, formData, (code, txt) => {
      callback && callback(code === 202, code === 202 ? `Succeed! ${txt}` : `Failed! ${txt}`);
    }, (msg) => {
      callback && callback(false, msg); 
    });
  },

  resetPasswordConfirm(username, formData, callback) {
    Fetch.text(`/api/users/${username}/reset-password`, 'POST', null, formData, (code, txt) => {
      if (code === 204) {
        callback && callback(true, `Succeed! Will redirect back to the homepage.`);
      } else {
        callback && callback(false, `Failed! Please try again later or reset your password again.`);
      }
    }, (msg) => {
      callback && callback(false, msg);
    });
  },

  register(formData, callback) {
    Fetch.json('/api/users', 'POST', null, formData, (code, data) => {
      if (code === 202) {
        callback && callback(true, `Succeed! ${data.message || data}`);
      } else {
        callback && callback(false, `Failed! ${data.error || data.message || 'Server got hacked'}`);
      }
    }, (msg) => {
      callback && callback(false, msg); 
    });
  },

};

export const kCookieToken = 'SSO_Site_Access';

export let Admin = {
  clientId: 1,
  realm: 'SSO-Site',
  secret: 'admin',
  
  listMembers(group, token, tokenType, callback) {
    Fetch.json(`/api/groups/${group}`, 'GET', {token, tokenType}, null, (code, data) => {
      callback && callback(code === 200 ? data.members : []);
    }, (msg) => {
      callback && callback([]); 
    });
  },

  listGroupMembers(group, token, tokenType, callback) {
    Fetch.json(`/api/groups/${group}`, 'GET', {token, tokenType}, null, (code, data) => {
      callback && callback(code === 200 ? data.group_members : []);
    }, (msg) => {
      callback && callback([]); 
    });
  },

  addMember(group, token, tokenType, formData, callback) {
    Fetch.text(`/api/groups/${group}/members/${formData.username}`, 'PUT', {token, tokenType}, formData, (code, txt) => {
      callback && callback(code === 200, code === 200 ? "Successfully add the user into group" : `Failed to add the user! ${txt}`);
    }, (msg) => {
      callback && callback(false, msg); 
    });
  },

  addGroupMember(group, token, tokenType, formData, callback) {
    Fetch.text(`/api/groups/${group}/group-members/${formData.sonname}`, 'PUT', {token, tokenType}, formData, (code, txt) => {
      callback && callback(code === 200, code === 200 ? "Successfully add the sub group into group" : `Failed to add the group! ${txt}`);
    }, (msg) => {
      callback && callback(false, msg); 
    });
  },


  deleteMember(group, token, tokenType, name, callback) {
    Fetch.text(`/api/groups/${group}/members/${name}`, 'DELETE', {token, tokenType}, null, (code, txt) => {
      callback && callback(code === 204, code === 204 ? "Successfully remove the user" : `Failed to remove the user! ${txt}`);
    }, (msg) => {
      callback && callback(false, msg);
    });
  },

  deleteGroupMember(group, token, tokenType, name, callback) {
    Fetch.text(`/api/groups/${group}/group-members/${name}`, 'DELETE', {token, tokenType}, null, (code, txt) => {
      callback && callback(code === 204, code === 204 ? "Successfully remove the group" : `Failed to remove the group! ${txt}`);
    }, (msg) => {
      callback && callback(false, msg);
    });
  },

  createGroup(token, tokenType, formData, callback) {
    Fetch.text('/api/groups', 'POST', {token, tokenType}, formData, (code, txt) => {
      if (code === 201) {
        callback && callback(true, 'Successfully create a group.'); 
      } else if (code === 409) {
        callback && callback(false, 'Same named group exists.');
      } else {
        callback && callback(false, `Failed to create the group, ${txt}`);
      }
    }, (msg) => {
      callback && callback(false, msg); 
    });
  },

  deleteGroup(token, tokenType, name, callback) {
    Fetch.text(`/api/groups/${name}`, 'DELETE', {token, tokenType}, null, (code, txt) => {
      callback && callback(code === 204, code === 204 ? "Successfully deleted the group" : txt);
    }, (msg) => {
      callback && callback(false, msg); 
    });
  },

  listGroups(token, tokenType, callback) {
    Fetch.json('/api/groups', 'GET', {token, tokenType}, null, (code, data) => {
      callback && callback(code === 200 ? data : []);
    }, (msg) => {
      callback && callback([]); 
    });
  },

  listUsers(token, tokenType, callback) {
    Fetch.json('/api/users', 'GET', {token, tokenType}, null, (code, data) => {
      callback && callback(code === 200 ? data : []);
    }, (msg) => {
      callback && callback([]); 
    });
  },

  listInactiveUsers(token, tokenType, callback) {
    Fetch.json('/api/inactiveusers', 'GET', {token, tokenType}, null, (code, data) => {
      callback && callback(code === 200 ? data : []);
    }, (msg) => {
      callback && callback([]); 
    });
  },

  activateUser(activationCode, callback) {
    Fetch.text(`/api/activateuser?code=${activationCode}`, 'GET', null, null, (code, txt) => {
      callback && callback(code === 201, code === 201 ? "Successfully activated the user" : txt);
    }, (msg) => {
      callback && callback(false, msg); 
    });
  },

  deleteUser(token, tokenType, username, callback) {
    Fetch.text(`/api/users/${username}`, 'DELETE', {token, tokenType}, null, (code, txt) => {
      callback && callback(code === 204, code === 204 ? "Successfully deleted the user" : txt);
    }, (msg) => {
      callback && callback(false, msg); 
    });
  },

  listApplications(token, tokenType, callback) {
    Fetch.json('/api/apps', 'GET', {token, tokenType}, null, (code, data) => {
      callback && callback(code === 200 ? data : []);
    }, (msg) => {
      callback && callback([]); 
    });
  },

  createApplication(token, tokenType, formData, callback) {
    Fetch.json('/api/apps', 'POST', {token, tokenType}, formData, (code, data) => {
      callback && callback(code === 201, JSON.stringify(data, null, '  '));
    }, (msg) => {
      callback && callback(false, msg); 
    });
  },

  isCurrUserAdmin(group, token, tokenType, callback) {
    Fetch.json('/api/me', 'GET', {token, tokenType}, null, (code, data) => {
      let isAdmin = false;
      if (code === 200) {
        isAdmin = _.some(data.groups, (g) => g === group); 
      }
      callback && callback(isAdmin);
    }, (msg) => {
      callback && callback(false); 
    });
  },

  setTokenCookie(token, tokenType, expires) {
    const tokenValue = {
      access: token,
      ty: tokenType,
    };
    Cookie.set(kCookieToken, tokenValue, {
      path: '/',
    });
  },

  getTokenCookie() {
    const tokens = Cookie.get(kCookieToken);
    if (tokens) {
      try {
        return JSON.parse(tokens);
      } catch (ex) {}  
    }
    return { access: '', ty: '' };
  },

  getToken(authCode, adminArea, checkGroup, callback) {
    let query = { area: adminArea };
    if (checkGroup) {
      query['ag'] = checkGroup;
    }
    const redirectUrl = `${window.location.protocol}//${window.location.host}/spa/admin/authorize?${this.toQuery(query)}`;
    let formData = {
      client_id: this.clientId,
      client_secret: this.secret,
      code: authCode,
      grant_type: 'authorization_code',
      redirect_uri: redirectUrl,
    };
    Fetch.json(`/oauth2/token?${this.toQuery(formData)}`, 'GET', null, null, (code, data) => {
      if (code === 200) {
        this.setTokenCookie(data.access_token, data.token_type, data.expires_in);
        callback && callback(true, data.access_token, data.token_type);
      } else {
        callback && callback(false);
      }
    }, (msg) => {
      callback && callback(false); 
    });
  },

  redirectOauth(adminArea, checkGroup) {
    let query = { area: adminArea };
    if (checkGroup) {
      query['ag'] = checkGroup;
    }
    const redirectUrl = `${window.location.protocol}//${window.location.host}/spa/admin/authorize?${this.toQuery(query)}`;
    let params = {
      response_type: 'code',
      redirect_uri: redirectUrl,
      realm: this.realm,
      client_id: this.clientId,
      scope: 'write:app read:app read:user write:user write:group read:group',
      state: Math.random(),
    };
    const oauthUrl = `${window.location.protocol}//${window.location.host}/oauth2/auth?${this.toQuery(params)}`;
    window.location.href = oauthUrl;
  },

  toQuery(kv) {
    let params = [];
    _.forOwn(kv, (value, key) => {
      params.push(`${key}=${encodeURIComponent(value)}`);
    });
    return params.join("&");
  },

};

export let Query={
  getUserInfo(name, callback){
    Fetch.json(`/api/users/${name}`, 'GET', null, null, (code, data) => {
      callback && callback( code === 200 ? this.getGroupAndRole(data, callback) : [] );
    }, (msg) => {
      callback && callback([]); 
    });
  },

  getGroupAndRole(data, callback){
    let ret = new Array(data.groups.length);
    for (let i=0; i < data.groups.length; i++){
      let group = data.groups[i];
      ret[i] = new Object;
      ret[i].name = group;
      Fetch.json(`/api/groups/${group}`, 'GET', null, null, (code1, data1) => {
        if (code1 === 200){
          ret[i].fullname=data1.fullname;
          callback(ret);
        }
      }, (msg) => {false});
      Fetch.json(`/api/groups/${group}/members/${data.name}`, 'GET', null, null, (code2, data2) => {
        if (code2 === 200){
          ret[i].userRole=data2.role;
          callback(ret);
        }
      }, (msg) => {false}); 
    };
    return ret;
  },
};

let Fetch = {
  
  text(api, method, auth, payload, succCb, errCb) {
    let code = 200;
    let headers = {
      'Accept': 'application/json',
      'Content-Type': 'application/json'
    };
    if (auth && auth.token && auth.tokenType) {
      const {token, tokenType} = auth;
      headers['Authorization'] = `${_.capitalize(tokenType)} ${token}`;
    }
    let options = { method, headers };
    if (payload) {
      options['body'] = JSON.stringify(payload);
    }
    fetch(api, options).then(response => {
      code = response.status;
      return response.text();
    }).then(txt => {
      succCb && succCb(code, txt); 
    }).catch(err => {
      console.log(`Error when ${method} ${api} ${payload}: ${err}`);
      errCb && errCb(`Server got hacked, ${err}`)
    });
  },

  json(api, method, auth, payload, succCb, errCb) {
    this.text(api, method, auth, payload, (code, txt) => {
      succCb && succCb(code, JSON.parse(txt));
    }, (msg) => {
      errCb && errCb(msg);
    })
  },

};

