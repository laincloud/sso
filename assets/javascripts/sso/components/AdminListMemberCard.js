import StyleSheet from 'react-style';
import React from 'react';

import {Admin} from '../models/Models';

let AdminListMemberCard = React.createClass({

  getInitialState() {
    return {
      members: [],
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
          <h2 className="mdl-card__title-text">{group}组用户列表</h2>
        </div>

        <table className="mdl-data-table mdl-js-data-table" style={this.styles.table}>
          <thead>
            <tr>
              <th className="mdl-data-table__cell--non-numeric">用户名</th>
              <th className="mdl-data-table__cell--non-numeric">身份</th>
              <th className="mdl-data-table__cell--non-numeric"></th>
            </tr>
          </thead>
          <tbody>
            {
              this.state.members.map((user, index) => {
                return (
                  <tr key={`member-${index}`}>
                    <td className="mdl-data-table__cell--non-numeric">{user.name}</td>
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
    let yes = confirm(`确定要踢掉用户 - ${name} 吗？`);
    if (yes) {
      const {group, token, tokenType} = this.props;
      Admin.deleteMember(group, token, tokenType, name, (ok, status) => ok ? this.reload() : alert(status));
    }
  },

  reload() {
    const {token, tokenType, group} = this.props; 
    Admin.listMembers(group, token, tokenType, (members) => {
      this.setState({ members });
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

export default AdminListMemberCard;
