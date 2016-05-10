import StyleSheet from 'react-style';
import React from 'react';

import {User} from '../models/Models';
import CardFormMixin from './CardFormMixin';

let UserResetPasswordCard = React.createClass({
  mixins: [CardFormMixin],

  getInitialState() {
    return {
      formValids: {
        'name': true,
      },
    };
  },

  componentDidUpdate() {
    componentHandler.upgradeDom();
  },

  render() {
    const {reqResult} = this.state;
    return (
      <div className="mdl-card mdl-shadow--2dp" styles={[this.styles.card, this.props.style]}>
        <div className="mdl-card__title">
          <h2 className="mdl-card__title-text">重置密码</h2>
        </div>
        { this.renderResult() }
        {
          reqResult.fin && reqResult.ok ? null :
            this.renderForm(this.onReset, [
              this.renderInput("name", "用户名*(字母、数字和减号)", { type: "text", pattern: "[\-a-zA-Z0-9]*" }), 
            ])
        }
        { this.renderAction("重置密码", this.onReset) }
      </div>
    );
  },

  onReset() {
    const {isValid, formData} = this.validateForm(["name"], ["name"]);
    if (isValid) {
      this.setState({ inRequest: true });
      User.resetPassword(formData['name'], this.onRequestCallback);
    }
  },

});

export default UserResetPasswordCard;
