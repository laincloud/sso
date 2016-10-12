import StyleSheet from 'react-style';
import React from 'react';
import {History} from 'react-router';

import {Admin} from '../models/Models';

let AdminListUserCard = React.createClass({
  mixins: [History],

  getInitialState() {
    return {
      users: [], 
    };
  },

  componentDidMount() {
    this.reload(); 
  },

  render() {
    return (
      <div className="mdl-card mdl-shadow--2dp" styles={[this.styles.card, this.props.style]}>
        <div className="mdl-card__title">
          <h2 className="mdl-card__title-text">用户列表</h2>
        </div>

        <table className="mdl-data-table mdl-js-data-table" style={this.styles.table}>
          <thead>
            <tr>
              <th className="mdl-data-table__cell--non-numeric">用户名</th>
              <th className="mdl-data-table__cell--non-numeric">Email</th>
              <th className="mdl-data-table__cell--non-numeric">用户组</th>
              <th className="mdl-data-table__cell--non-numeric"></th>
            </tr>
          </thead>
          <tbody>
            {
              this.state.users.map((user, index) => {
                return (
                  <tr key={`user-${index}`}>
                    <td className="mdl-data-table__cell--non-numeric">{user.name}</td>
                    <td className="mdl-data-table__cell--non-numeric">{user.email}</td>
                    <td className="mdl-data-table__cell--non-numeric" style={this.styles.breaklineTd}>
                      {
                        _.map(user.groups, (group) => {
                          return (
                            <a href="javascript:;" key={`group-${group}`} 
                              style={{ marginRight: 8 }}
                              onClick={(evt) => this.goGroupDetail(group)}>{group}</a> 
                          );
                        })
                      }
                    </td>
                    <td className="mdl-data-table__cell--non-numeric">
                      {
                        user.name === 'admin' ? null : 
                          <a href="javascript:;" onClick={(evt) => this.deleteUser(user.name)}>删除</a>
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

  deleteUser(username) {
    let yes = confirm(`确定要删除用户 - ${username} 吗？`);
    if (yes) {
      const {token, tokenType} = this.props;
      Admin.deleteUser(token, tokenType, username, (ok, status) => {
        if (ok) {
          this.reload(); 
        }
      });
    }
  },

  reload() {
    const {token, tokenType} = this.props;
    Admin.listUsers(token, tokenType, (users) => {
      this.setState({ users });
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
      position: 'relative',
      'vertical-align': 'top',
      height: '48px',
      'border-top': '1px solid rgba(0,0,0,.12)',
      'border-bottom': '1px solid rgba(0,0,0,.12)',
      padding: '12px 18px 0',
      'box-sizing': 'border-box',
      'white-space': 'pre-line',
      'word-wrap': 'break-word',
      'table-layout': 'fixed',
      'word-break': 'break-all',
    },
  }),
});

export default AdminListUserCard;
