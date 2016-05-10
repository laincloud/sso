import StyleSheet from 'react-style';
import React from 'react';
import {History} from 'react-router';

import AdminNewAppCard from '../components/AdminNewAppCard';
import AdminListAppCard from '../components/AdminListAppCard';
import AdminAuthorizeMixin from '../components/AdminAuthorizeMixin';
import {Admin} from '../models/Models';

let AdminAppsPage = React.createClass({
  mixins: [History, AdminAuthorizeMixin],

  componentWillMount() {
    this.authorize('apps');
  },

  render() {
    const isValid = this.isSessionValid();
    return (
      <div className="mdl-grid">
        <div className="mdl-cell mdl-cell--6-col mdl-cell--8-col-tablet mdl-cell--4-col-phone">
          { 
            !isValid ? <p>等待认证中....</p>
              : <AdminNewAppCard token={this.state.token} tokenType={this.state.tokenType} onSucc={this.postCreateApp} />
          }
        </div>

        <div className="mdl-cell mdl-cell--12-col mdl-cell--8-col-tablet mdl-cell--4-col-phone">
          {
            !isValid ? <p>等待认证中……</p>
              : <AdminListAppCard ref="appList" token={this.state.token} tokenType={this.state.tokenType} />
          }
        </div>
      </div>
    );
  },

  postCreateApp() {
    if (this.isSessionValid()) {
      this.refs.appList.reload();
    }
  },
  
  styles: StyleSheet.create({
  }),

});

export default AdminAppsPage;
