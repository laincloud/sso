import StyleSheet from 'react-style';
import React from 'react';
import {History} from 'react-router';

import AdminNewGroupCard from '../components/AdminNewGroupCard';
import AdminDeleteGroupCard from '../components/AdminDeleteGroupCard';
import AdminListGroupCard from '../components/AdminListGroupCard';
import AdminAuthorizeMixin from '../components/AdminAuthorizeMixin';
import {Admin} from '../models/Models';

let AdminGroupsPage = React.createClass({
  mixins: [History, AdminAuthorizeMixin],

  componentWillMount() {
    this.authorize('groups');
  },

  componentDidUpdate(){
    this.authorize('groups');
  },

  render() {
    const isValid = this.isSessionValid();
    return (
      <div className="mdl-grid">
        <div className="mdl-cell mdl-cell--4-col mdl-cell--8-col-tablet mdl-cell--4-col-phone">
          { 
            !isValid ? <p>等待认证中....</p>
              : <AdminNewGroupCard token={this.state.token} tokenType={this.state.tokenType} onSucc={this.refreshGroups} />
          }
        </div>
        <div className="mdl-cell mdl-cell--12-col mdl-cell--8-col-tablet mdl-cell--4-col-phone">
          {
            !isValid ? <p>等待认证中……</p>
              : <AdminListGroupCard ref="groupList" token={this.state.token} tokenType={this.state.tokenType} />
          }
        </div>
      </div>
    );
  },

  refreshGroups() {
    if (this.isSessionValid()) {
      this.refs.groupList.reload();
    }
  },

  styles: StyleSheet.create({
  }),

});

export default AdminGroupsPage;
