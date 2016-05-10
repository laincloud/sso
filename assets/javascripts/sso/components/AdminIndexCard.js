import StyleSheet from 'react-style';
import React from 'react';
import {History} from 'react-router';

let AdminIndexCard = React.createClass({
  mixins: [History],
  
  render() {
    const buttons = [
      { title: "我的应用管理", target: "apps" },
      { title: "我的群组管理", target: "groups" },
      { title: "用户管理－管理员特供", target: "users" },
    ];
    
    return (
      <div className="mdl-card mdl-shadow--2dp" styles={[this.styles.card, this.props.style]}>
        <div className="mdl-card__title">
          <h2 className="mdl-card__title-text">自助服务</h2>
        </div>
        <div className="mdl-card__supporting-text" style={this.styles.supporting}>
          这里提供了一些应用和群组的管理功能，用户管理属于管理员特供功能，非管理员同学请勿操作，如果有需要，请到C座21层平台组喊一声。
        </div>
        
        <div style={{ padding: 8 }}>
          {
            _.map(buttons, (btn) => {
              return (
                <button className="mdl-button mdl-js-button mdl-button--accent mdl-js-ripple-effect"
                  onClick={(evt) => this.adminAuthorize(btn.target)}
                  key={btn.target}
                  style={this.styles.buttons}>
                  {btn.title}
                </button>
              ); 
            })
          }
        </div>
      </div>
    );  
  },

  adminAuthorize(target) {
    this.history.pushState(null, `/spa/admin/${target}`);
  },

  styles: StyleSheet.create({
    card: {
      width: '100%',
      marginBottom: 16,
      minHeight: 50,
    },
    buttons: {
      display: 'block',
    },
    supporting: {
      borderTop: '1px solid rgba(0, 0, 0, .12)',
      borderBottom: '1px solid rgba(0, 0, 0, .12)',
    },
  }),

});

export default AdminIndexCard;
