import StyleSheet from 'react-style';
import React from 'react';

import {User} from '../models/Models';
import CardFormMixin from './CardFormMixin';

let UserResetPasswordConfirmCard = React.createClass({
  mixins: [CardFormMixin],

  getInitialState() {
    return {
      formValids: {
        'password': true,
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
          <h2 className="mdl-card__title-text">设置新密码</h2>
        </div>
        { this.renderResult() }
        {
          reqResult.fin && reqResult.ok ? null :
            this.renderForm(this.onReset, [
              this.renderInput("password", "新密码*", { type: 'password' }),
            ])
        }
        { this.renderAction("设置新密码", this.onReset) }
      </div>
    );
  },

  onReset() {
    const {isValid, formData} = this.validateForm(["password"], ["password"]);
    const {code, username} = this.props;
    if (isValid) {
      this.setState({ inRequest: true });
      formData['code'] = code;
      User.resetPasswordConfirm(username, formData, (ok, status) => {
        this.setState({
          inRequest: false,
          reqResult: { fin: true, ok, status},
        }, () => {
          if (ok) {
            this.props.onConfirm && this.props.onConfirm();
          }
        });
      });
    }
  },

});

export default UserResetPasswordConfirmCard;
