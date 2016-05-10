import React from 'react';
import {History} from 'react-router';

import {Admin} from '../models/Models';

let AdminAuthorizePage = React.createClass({
  mixins: [History],

  getInitialState() {
    return {
      message: '认证中',
    };
  },
  
  componentDidMount() {
    const {area, code, ag} = this.props.location.query;
    if (!area) {
      this.history.replaceState(null, '/spa/');
      return
    }

    if (code) {
      Admin.getToken(code, area, ag, (ok, token, tokenType) => {
        if (ok) {
          this.redirectWithToken(area, ag, token, tokenType);
        } else {
          this.setStatusAndBounce("获取Token失败～")
        }
      });
    } else {
      const {access, ty} = Admin.getTokenCookie();
      if (access && ty) {
        this.redirectWithToken(area, ag, access, ty);
      } else {
        Admin.redirectOauth(area, ag);
      }
    }
  },

  redirectWithToken(area, ag, token, tokenType) {
    if (ag) {
      this.checkAdminUser(area, ag, token, tokenType);
    } else {
      this.redirectArea(area, token, tokenType);
    }
  },

  redirectArea(area, token, tokenType) {
    this.history.replaceState({
      token, 
      tokenType,
    }, `/spa/admin/${area}`);
  },

  checkAdminUser(area, group, token, tokenType) {
    Admin.isCurrUserAdmin(group, token, tokenType, (ok) => {
      if (ok) {
        this.redirectArea(area, token, tokenType);
      } else {
        this.setStatusAndBounce('貌似您不是Admin管理员哦～');
      }
    }); 
  },

  setStatusAndBounce(message) {
    this.setState({ message }, () => {
      setTimeout(() => {
        this.history.replaceState(null, '/spa/');
      }, 3000);
    })
  },

  componentWillUpdate() {
    componentHandler.upgradeDom();
  },

  render() {
    return (
      <div style={{ marginTop: 48 }}>
        <p>{this.state.message} ……</p>
        <div className="mdl-progress mdl-js-progress mdl-progress__indeterminate"></div>
      </div>
    );
  },

});

export default AdminAuthorizePage;
