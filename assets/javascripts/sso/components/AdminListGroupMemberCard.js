import StyleSheet from 'react-style';
import React from 'react';
import {History} from 'react-router';

import {Admin} from '../models/Models';

let AdminListGroupMemberCard = React.createClass({

  mixins: [History],

  getInitialState() {
    return {
      group_members: [],
    };
  },

  componentWillMount() {
    this.reload();
  },

  componentWillReceiveProps(nextProps){
    const {token, tokenType, group} = nextProps;
    const group_members = this.state.group_members;
    Admin.listGroupMembers(group, token, tokenType, (group_members) => {
      this.setState({ group_members });
    });
  },

  render() {
    const {group} = this.props;
    return (
      <div className="mdl-card mdl-shadow--2dp" styles={[this.styles.card, this.props.style]}>
        <div className="mdl-card__title">
          <h2 className="mdl-card__title-text">{group}子组列表</h2>
        </div>

        <table className="mdl-data-table mdl-js-data-table" style={this.styles.table}>
          <thead>
            <tr>
			  <th className="mdl-data-table__cell--non-numeric">组名</th>
			  <th className="mdl-data-table__cell--non-numeric">描述</th>
			 <th className="mdl-data-table__cell--non-numeric">身份</th>
              <th className="mdl-data-table__cell--non-numeric"></th>
            </tr>
          </thead>
          <tbody>
            {
              this.state.group_members.map((subgroup, index) => {
                return (
                  <tr key={`gmember-${index}`}>
                    <td className="mdl-data-table__cell--non-numeric" style={this.styles.breaklineTd}>
                      {
                        <a href="javascript:;" key={`subgroup-${subgroup.name}`} 
                          style={{ marginRight: 8 }}
                          onClick={(evt) => this.goGroupDetail(subgroup.name)}>{subgroup.name}</a>
                      }
                    </td>
					  <td className="mdl-data-table__cell--non-numeric" style={this.styles.breaklineTd}>{subgroup.fullname}</td>

                    <td className="mdl-data-table__cell--non-numeric">{subgroup.role === "admin" ? "管理员" : "成员"}</td>
                    <td className="mdl-data-table__cell--non-numeric">
                      {
                        subgroup.role === 'admin' ? null : 
                          <a href="javascript:;" onClick={(evt) => this.deleteMember(subgroup.name)}>删除</a>
                      }
                    </td>
                  </tr>
                );
              })
            }
          </tbody>
        </table>

        <div className="mdl-card__actions">
          <button className="mdl-button mdl-js-button mdl-js-ripple-effect mdl-button--colored"
            onClick={this.reload}>刷新</button>
        </div>
      </div>
    );
  },

  goGroupDetail(name) {
    const {token, tokenType} = this.props;
    this.history.pushState({token, tokenType}, `/spa/admin/groups/${name}`);
  },

  deleteMember(name) {
    let yes = confirm(`确定要踢掉子组 - ${name} 吗？`);
    if (yes) {
      const {group, token, tokenType} = this.props;
      Admin.deleteGroupMember(group, token, tokenType, name, (ok, status) => ok ? this.reload() : alert(status));
    }
  },

  reload() {
    const {token, tokenType, group} = this.props; 
    Admin.listGroupMembers(group, token, tokenType, (group_members) => {
      this.setState({ group_members });
    });
  },

  styles: StyleSheet.create({
    card: {
      width: '100%',
      marginBottom: 16,
      minHeight: 50,
    },
    table: {
      width: '100%',
      borderLeft: 'none',
      borderRight: 'none',
    },
    breaklineTd: {
      whiteSpace: 'pre-line',
      wordWrap: 'break-word',
      wordBreak: 'break-word',
    },
  }),
});

export default AdminListGroupMemberCard;
