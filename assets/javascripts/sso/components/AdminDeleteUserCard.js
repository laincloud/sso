import StyleSheet from 'react-style';
import React from 'react';

import CardFormMixin from './CardFormMixin';
import {Admin} from '../models/Models';

let AdminDeleteUserCard = React.createClass({
  mixins: [CardFormMixin],

  getInitialState() {
    return {
      formValids: {
        'username': true,
      },
    };
  },

  render() {
    const {reqResult} = this.state;
    return (
      <div className="mdl-card mdl-shadow--2dp" styles={[this.styles.card, this.props.style]}>
        <div className="mdl-card__title">
          <h2 className="mdl-card__title-text">删除用户</h2>
        </div>
        { this.renderResult() }
        { 
          reqResult.fin && reqResult.ok ? null :
            this.renderForm(this.onDelete, [
              this.renderInput("username", "用户名*(字母、数字和减号)", { type: "text", pattern: "[\-a-zA-Z0-9]*" }),
            ])
        }
        { this.renderAction("确定删除", this.onDelete) }
      </div>
    );
  },

  onDelete() {
    const fields = ['username'];
    const {isValid, formData} = this.validateForm(fields, fields);
    if (isValid) {
      const {token, tokenType} = this.props;
      this.setState({ inRequest: true });
      Admin.deleteUser(token, tokenType, formData.username, this.onRequestCallback);
    }
  },

});

export default AdminDeleteUserCard;
