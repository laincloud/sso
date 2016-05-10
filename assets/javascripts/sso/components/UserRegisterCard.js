import StyleSheet from 'react-style';
import React from 'react';

import {User} from '../models/Models';
import CardFormMixin from './CardFormMixin';

let UserRegisterCard = React.createClass({
  mixins: [CardFormMixin],

  getInitialState() {
    return {
      formValids: {
        'name': true,
        'email': true,
        'password': true,
      },
    };
  },

  render() {
    const {reqResult} = this.state;
    return (
      <div className="mdl-card mdl-shadow--2dp" styles={[this.styles.card, this.props.style]}>
        <div className="mdl-card__title">
          <h2 className="mdl-card__title-text">新用户注册</h2>
        </div>
        { this.renderResult() }
        { 
          reqResult.fin && reqResult.ok ? null :
            this.renderForm(this.onRegister, [
              this.renderInput("name", "用户名*(字母、数字和减号)", { type: "text", pattern: "[\-a-zA-Z0-9]*" }),
              this.renderInput("fullname", "全名", { type: 'text' }),
              this.renderInput("email", "Email*(仅限公司Email地址)", { type: 'email' }),
              this.renderInput("password", "密码*", { type: 'password' }),
              this.renderInput("mobile", "手机号", { type: 'tel' }),
            ])
        }
        { this.renderAction("提交注册", this.onRegister) }
      </div>
    );
  },

  onRegister() {
    const fields = ['name', 'fullname', 'email', 'password', 'mobile'];
    const rFields = ['name', 'email', 'password'];
    const {isValid, formData} = this.validateForm(fields, rFields);
    if (isValid) {
      this.setState({ inRequest: true });
      User.register(formData, this.onRequestCallback);
    }
  },

});

export default UserRegisterCard;
