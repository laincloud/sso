import StyleSheet from 'react-style';
import React from 'react';
import {History} from 'react-router';

import {Admin} from '../models/Models';

let AdminListGroupCard = React.createClass({
  mixins: [History],

  getInitialState() {
    return {
      groups: [], 
    };
  },

  componentDidMount() {
    this.reload(); 
  },

  render() {
    return (
      <div className="mdl-card mdl-shadow--2dp" styles={[this.styles.card, this.props.style]}>
        <div className="mdl-card__title">
          <h2 className="mdl-card__title-text">我的用户组列表</h2>
        </div>

        <table className="mdl-data-table mdl-js-data-table" style={this.styles.table}>
          <thead>
            <tr>
              <th className="mdl-data-table__cell--non-numeric">名称</th>
              <th className="mdl-data-table__cell--non-numeric">描述</th>
              <th className="mdl-data-table__cell--non-numeric">身份</th>
              <th className="mdl-data-table__cell--non-numeric"></th>
            </tr>
          </thead>
          <tbody>
            {
              this.state.groups.map((group, index) => {
                return (
                  <tr key={`group-${index}`}>
                    <td className="mdl-data-table__cell--non-numeric">
                      <a href="javascript:;" onClick={(evt) => this.goDetail(group.name)}>{group.name}</a>
                    </td>
                    <td className="mdl-data-table__cell--non-numeric" style={this.styles.breaklineTd}>{group.fullname}</td>
                    <td className="mdl-data-table__cell--non-numeric">{group.role === "admin" ? "管理员" : "成员"}</td>
                    <td className="mdl-data-table__cell--non-numeric">
                      {
                        !this.canDelete(group) ? null : 
                          <a href="javascript:;" onClick={(evt) => this.deleteGroup(group.name)}>删除</a>
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

  goDetail(name) {
    const {token, tokenType} = this.props;
    this.history.pushState({token, tokenType}, `/spa/admin/groups/${name}`);
  },

  canDelete(group) {
    return !(group.name === 'admins' || group.role !== 'admin' || _.startsWith(group.name, "."));
  },

  deleteGroup(name) {
    let yes = confirm(`确定要删除用户组 - ${name} 吗？`);
    if (yes) {
      const {token, tokenType} = this.props;
      Admin.deleteGroup(token, tokenType, name, (ok, status) => ok ? this.reload() : null);
    }
  },

  reload() {
    const {token, tokenType} = this.props;
    Admin.listGroups(token, tokenType, (groups) => {
      this.setState({ groups });
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

export default AdminListGroupCard;
