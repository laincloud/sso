import StyleSheet from 'react-style';
import React from 'react';
import {History} from 'react-router';

import AdminAuthorizeMixin from '../components/AdminAuthorizeMixin';
import AdminNewMemberCard from '../components/AdminNewMemberCard';
import AdminListMemberCard from '../components/AdminListMemberCard';
import AdminNewGroupMemberCard from '../components/AdminNewGroupMemberCard';
import AdminListGroupMemberCard from '../components/AdminListGroupMemberCard';
import {Admin} from '../models/Models';

let AdminMembersPage = React.createClass({
  mixins: [History, AdminAuthorizeMixin],

  componentWillMount() {
    this.authorize(`groups/${this.getGroupName()}`);
  },

  componentDidUpdate() {
    this.authorize(`groups/${this.getGroupName()}`);
  },

  render() {
    const isValid = this.isSessionValid();
    const groupName = this.getGroupName();
    return (
      <div className="mdl-grid">
        <div className="mdl-cell mdl-cell--8-col mdl-cell--8-col-tablet mdl-cell--4-col-phone">
          {
            !isValid ? <p>等待认证中……</p> :
			  [
				<AdminListMemberCard ref="memberList" token={this.state.token} tokenType={this.state.tokenType}
				  group={groupName} />,
                <AdminListGroupMemberCard ref="groupMemberList" token={this.state.token} tokenType={this.state.tokenType}
				  group={groupName} />
			  ]
         }
        </div>
        <div className="mdl-cell mdl-cell--4-col mdl-cell--8-col-tablet mdl-cell--4-col-phone">
          {
            !isValid ? <p>等待认证中……</p> : 
             [ <AdminNewMemberCard token={this.state.token} tokenType={this.state.tokenType} 
				 group={groupName} onSucc={this.refresh} />,
              <AdminNewGroupMemberCard token={this.state.token} tokenType={this.state.tokenType} 
				 group={groupName} onSucc={this.refresh} />]
}
        </div>
      </div>
    );
  },

  getGroupName() {
    const {params} = this.props;
    return params ? params.name : '';
  },

  refresh() {
    if (this.isSessionValid()) {
		this.refs.memberList.reload();
		this.refs.groupMemberList.reload();
    } 
  },

});

export default AdminMembersPage;
