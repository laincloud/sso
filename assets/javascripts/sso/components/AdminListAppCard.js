import StyleSheet from 'react-style';
import React from 'react';
import {History} from 'react-router';

import {Admin} from '../models/Models';

let AdminListAppCard = React.createClass({

  mixins: [History],

  getInitialState() {
    return {
      apps: [], 
    };
  },

  componentDidMount() {
    this.reload(); 
  },

  render() {
    return (
      <div className="mdl-card mdl-shadow--2dp" styles={[this.styles.card, this.props.style]}>
        <div className="mdl-card__title">
          <h2 className="mdl-card__title-text">我的应用列表</h2>
        </div>

        <table className="mdl-data-table mdl-js-data-table" style={this.styles.table}>
          <thead>
            <tr>
              <th>ClientID</th>
              <th className="mdl-data-table__cell--non-numeric">名称</th>
              <th className="mdl-data-table__cell--non-numeric">秘密</th>
              <th className="mdl-data-table__cell--non-numeric">管理组</th>
              <th className="mdl-data-table__cell--non-numeric">回调URL</th>
            </tr>
          </thead>
          <tbody>
            {
              this.state.apps.map((app, index) => {
                return (
                  <tr key={`app-${index}`}>
                    <td>{app.id}</td>
                    <td className="mdl-data-table__cell--non-numeric">{app.fullname}</td>
                    <td className="mdl-data-table__cell--non-numeric">{app.secret}</td>
                    <td className="mdl-data-table__cell--non-numeric">
                      {
                        <a href="javascript:;" key={`group-${app.admin_group.name}`}
                          style={{ marginRight: 8 }}
                          onClick={(evt) => this.goGroupDetail(app.admin_group.name)}>{app.admin_group.name}</a>
                      }
                    </td>
                    <td className="mdl-data-table__cell--non-numeric">{app.redirect_uri}</td>
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

  reload() {
    const {token, tokenType} = this.props;
    Admin.listApplications(token, tokenType, (apps) => {
      this.setState({ apps });
    });  
  },

  goGroupDetail(name) {
    const {token, tokenType} = this.props;
    this.history.pushState({token, tokenType}, `/spa/admin/groups/${name}`);
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

export default AdminListAppCard;
