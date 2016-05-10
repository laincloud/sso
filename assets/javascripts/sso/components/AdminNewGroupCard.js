import StyleSheet from 'react-style';
import React from 'react';

import CardFormMixin from './CardFormMixin';
import {Admin} from '../models/Models';

let AdminNewGroupCard = React.createClass({
  mixins: [CardFormMixin],

  getInitialState() {
    return {
      formValids: {
        'name': true,
      },
    };
  },

  render() {
    const {reqResult} = this.state;
    return (
      <div className="mdl-card mdl-shadow--2dp" styles={[this.styles.card, this.props.style]}>
        <div className="mdl-card__title">
          <h2 className="mdl-card__title-text">添加用户组</h2>
        </div>
        { this.renderResult() }
        { 
          reqResult.fin && reqResult.ok ? null :
            this.renderForm(this.onCreate, [
              this.renderInput("name", "用户组名称*", { type: 'text' }),
              this.renderInput("fullname", "用户组全名描述", { type: 'text' }),
            ])
        }
        { this.renderAction("保存", this.onCreate) }
      </div>
    );
  },

  onCreate() {
    const {isValid, formData} = this.validateForm(['name', 'fullname'], ['name']);
    if (isValid) {
      const {token, tokenType} = this.props;
      this.setState({ inRequest: true });
      Admin.createGroup(token, tokenType, formData, this.onRequestCallback);
    }
  },
    
});

export default AdminNewGroupCard;
