import StyleSheet from 'react-style';
import React from 'react';

import CardFormMixin from './CardFormMixin';
import {Admin} from '../models/Models';

let AdminDeleteGroupCard = React.createClass({
  mixins: [CardFormMixin],

  getInitialState() {
    return {
      formValids: {
        'groupname': true,
      },
    };
  },

  render() {
    const {reqResult} = this.state;
    return (
      <div className="mdl-card mdl-shadow--2dp" styles={[this.styles.card, this.props.style]}>
        <div className="mdl-card__title">
          <h2 className="mdl-card__title-text">删除用户组</h2>
        </div>
        { this.renderResult() }
        { 
          reqResult.fin && reqResult.ok ? null :
            this.renderForm(this.onDelete, [
              this.renderInput("groupname", "用户组名称*", { type: 'text' }),
            ])
        }
        { this.renderAction("确认删除", this.onDelete) }
      </div>
    );
  },

  onDelete() {
    const fields = ['groupname'];
    const {isValid, formData} = this.validateForm(fields, fields);
    if (isValid) {
      const {token, tokenType} = this.props;
      this.setState({ inRequest: true });
      Admin.deleteGroup(token, tokenType, formData.groupname, this.onRequestCallback);
    }
  },
});

export default AdminDeleteGroupCard;
