import StyleSheet from 'react-style';
import React from 'react';
import {History} from 'react-router';

import {Admin} from '../models/Models';

let AdminListInactiveUsersCard = React.createClass({
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
          <h2 className="mdl-card__title-text">未激活用户列表</h2>
        </div>

        <table className="mdl-data-table mdl-js-data-table" style={this.styles.table}>
          <thead>
            <tr>
              <th className="mdl-data-table__cell--non-numeric">用户名</th>
              <th className="mdl-data-table__cell--non-numeric">Email</th>
              <th className="mdl-data-table__cell--non-numeric">ActivationCode</th>
              <th className="mdl-data-table__cell--non-numeric"></th>
            </tr>
          </thead>
          <tbody>
            {
              this.state.users.map((user, index) => {
                return (
                  <tr key={`user-${index}`}>
                    <td className="mdl-data-table__cell--non-numeric">{user.Name}</td>
                    <td className="mdl-data-table__cell--non-numeric">{user.Email['String']}</td>
                    <td className="mdl-data-table__cell--non-numeric">{user.ActivationCode}</td>
                    <td className="mdl-data-table__cell--non-numeric">
                      {
                        user.name === 'admin' ? null : 
                          <a href="javascript:;" onClick={(evt) => this.activateUser(user.ActivationCode)}>激活</a>
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

  activateUser(activationCode) {
    Admin.activateUser(activationCode, (ok, status) => {
      if (ok) {
        this.reload(); 
      }
    });
  },

  reload() {
    const {token, tokenType} = this.props;
    Admin.listInactiveUsers(token, tokenType, (users) => {
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
  }),
});

export default AdminListInactiveUsersCard;
