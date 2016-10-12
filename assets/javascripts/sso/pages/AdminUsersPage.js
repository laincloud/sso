import StyleSheet from 'react-style';
import React from 'react';
import {History} from 'react-router';

import AdminDeleteUserCard from '../components/AdminDeleteUserCard';
import AdminListUserCard from '../components/AdminListUserCard';
import AdminListInactiveUsersCard from '../components/AdminListInactiveUsersCard';
import AdminAuthorizeMixin from '../components/AdminAuthorizeMixin';
import {Admin} from '../models/Models';

let AdminUsersPage = React.createClass({
  mixins: [History, AdminAuthorizeMixin],

  componentWillMount() {
    this.authorize('users', 'admins');
  },

  render() {
    const isValid = this.isSessionValid();
    return (
      <div className="mdl-grid">
        <div className="mdl-cell mdl-cell--6-col mdl-cell--8-col-tablet mdl-cell--4-col-phone">
          { 
            !isValid ? <p>等待认证中....</p>
              : <AdminDeleteUserCard token={this.state.token} tokenType={this.state.tokenType} onSucc={this.postDeleteUser} />
          }
        </div>

        <div className="mdl-cell mdl-cell--12-col mdl-cell--8-col-tablet mdl-cell--4-col-phone">
          {
            !isValid ? <p>等待认证中……</p>
              : <AdminListUserCard ref="userList" token={this.state.token} tokenType={this.state.tokenType} />
          }
        </div>

        <div className="mdl-cell mdl-cell--12-col mdl-cell--8-col-tablet mdl-cell--4-col-phone">
          {
            !isValid ? <p>等待认证中……</p>
              : <AdminListInactiveUsersCard ref="userList" token={this.state.token} tokenType={this.state.tokenType} />
          }
        </div>

      </div>
    );
  },

  postDeleteUser() {
    if (this.isSessionValid()) {
      this.refs.userList.reload();
    }
  },

  styles: StyleSheet.create({
  }),

});

export default AdminUsersPage;
