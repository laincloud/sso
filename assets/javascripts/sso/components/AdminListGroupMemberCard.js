import StyleSheet from 'react-style';
import React from 'react';

import {Admin} from '../models/Models';

let AdminListGroupMemberCard = React.createClass({

  getInitialState() {
    return {
      group_members: [],
    };
  },

  componentWillMount() {
    this.reload();
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
              this.state.group_members.map((user, index) => {
                return (
                  <tr key={`gmember-${index}`}>
					  <td className="mdl-data-table__cell--non-numeric">{user.name}</td>
					  <td className="mdl-data-table__cell--non-numeric">{user.fullname}</td>

                    <td className="mdl-data-table__cell--non-numeric">{user.role === "admin" ? "管理员" : "成员"}</td>
                    <td className="mdl-data-table__cell--non-numeric">
                      {
                        user.role === 'admin' ? null : 
                          <a href="javascript:;" onClick={(evt) => this.deleteMember(user.name)}>删除</a>
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
  }),
});

export default AdminListGroupMemberCard;
