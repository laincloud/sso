import StyleSheet from 'react-style';
import React from 'react';
import {History} from 'react-router';

import {Query} from '../models/Models';

let UserPublicInfoCard = React.createClass({
  mixins: [History],

  getInitialState() {
    return {
      groups: [], 
    };
  },

  componentDidMount() {
    this.reload(); 
  },

  componentWillReceiveProps(nextProps){
    const p = nextProps;
    Query.getUserInfo(p.user, (groups) => {
      //groups.sort(function(a, b){
      //    return a.name.localeCompare(b.name);
      //});
      this.setState({ groups });
    });  
 },

  render() {
    return (
      <div className="mdl-card mdl-shadow--2dp" styles={[this.styles.card, this.props.style]}>
        <div className="mdl-card__title">
          <h2 className="mdl-card__title-text">用户所在组列表</h2>
        </div>

        <table className="mdl-data-table mdl-js-data-table" style={this.styles.table}>
          <thead>
            <tr>
              <th className="mdl-data-table__cell--non-numeric">组名</th>
              <th className="mdl-data-table__cell--non-numeric">组详细描述</th>
              <th className="mdl-data-table__cell--non-numeric">用户角色</th>
            </tr>
          </thead>
          <tbody>
            {
              this.state.groups.map((group, index) => {
                return (
                  <tr key={`group-${index}`}>
                    <td className="mdl-data-table__cell--non-numeric">
                      <a href="javascript:;" key={`group-${group}`} 
                        style={{ marginRight: 8 }}
                        onClick={(evt) => this.goGroupDetail(group.name)}>{group.name}</a> 
                    </td>
                    <td className="mdl-data-table__cell--non-numeric">{group.fullname}</td>
                    <td className="mdl-data-table__cell--non-numeric">
                      {
                        group.userRole
                      }
                    </td>
                  </tr>
                  );
              })
            }
          </tbody>
        </table>
      </div>
    );
  },

  goGroupDetail(name) {
    // 其实如果仅仅查询用户或组的信息不需要认证，为了利用之前的模块
    this.history.pushState(null, `/spa/admin/groups/${name}`);
  },

  reload() {
    const p = this.props;
    Query.getUserInfo(p.user, (groups) => {
      //groups.sort(function(a, b){
      //  return a.name.localeCompare(b.name);
      //});
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
  }),
});

export default UserPublicInfoCard;
