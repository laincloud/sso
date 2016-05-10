import StyleSheet from 'react-style';
import React from 'react';

import CardFormMixin from './CardFormMixin';
import {Admin} from '../models/Models';

let AdminNewMemberCard = React.createClass({
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
          <h2 className="mdl-card__title-text">添加组成员</h2>
        </div>
        { this.renderResult() }
        { 
          reqResult.fin && reqResult.ok ? null :
            this.renderForm(this.onCreate, [
              this.renderInput("username", "用户名*", { type: 'text' }),
              this.renderInput("role", "成员身份(admin/normal)", { type: 'text' }),
            ])
        }
        { this.renderAction("添加", this.onCreate) }
      </div>
    );
  },

  onCreate() {
    const fields = ['username', 'role'];
    const {isValid, formData} = this.validateForm(fields, ['username']);
    if (isValid) {
      const {token, tokenType, group} = this.props;
      this.setState({ inRequest: true });
      Admin.addMember(group, token, tokenType, formData, this.onRequestCallback);
    }
  },

});

export default AdminNewMemberCard;
