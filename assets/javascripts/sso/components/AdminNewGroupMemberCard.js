import StyleSheet from 'react-style';
import React from 'react';

import CardFormMixin from './CardFormMixin';
import {Admin} from '../models/Models';

let AdminNewGroupMemberCard = React.createClass({
  mixins: [CardFormMixin],

  getInitialState() {
    return {
      formValids: {
        'sonname': true,
      },
    };
  },

  render() {
    const {reqResult} = this.state;
    return (
      <div className="mdl-card mdl-shadow--2dp" styles={[this.styles.card, this.props.style]}>
        <div className="mdl-card__title">
          <h2 className="mdl-card__title-text">添加成员组</h2>
        </div>
        { this.renderResult() }
        { 
          reqResult.fin && reqResult.ok ? null :
            this.renderForm(this.onCreate, [
              this.renderInput("sonname", "子组名*", { type: 'text' }),
              this.renderInput("role", "子组身份(admin/normal)", { type: 'text' }),
            ])
        }
        { this.renderAction("添加", this.onCreate) }
      </div>
    );
  },

  onCreate() {
    const fields = ['sonname', 'role'];
    const {isValid, formData} = this.validateForm(fields, ['sonname']);
    if (isValid) {
      const {token, tokenType, group} = this.props;
      this.setState({ inRequest: true });
      Admin.addGroupMember(group, token, tokenType, formData, this.onRequestCallback);
    }
  },

});

export default AdminNewGroupMemberCard;
