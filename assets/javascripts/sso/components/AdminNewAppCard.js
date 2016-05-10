import StyleSheet from 'react-style';
import React from 'react';

import CardFormMixin from './CardFormMixin';
import {Admin} from '../models/Models';

let AdminNewAppCard = React.createClass({
  mixins: [CardFormMixin],

  getInitialState() {
    return {
      formValids: {
        'fullname': true,
        'redirect_uri': true,
      },
    };
  },

  render() {
    const {reqResult} = this.state;
    return (
      <div className="mdl-card mdl-shadow--2dp" styles={[this.styles.card, this.props.style]}>
        <div className="mdl-card__title">
          <h2 className="mdl-card__title-text">添加客户端/应用</h2>
        </div>
        { this.renderResult() }
        { 
          reqResult.fin && reqResult.ok ? null :
            this.renderForm(this.onCreate, [
              this.renderInput("fullname", "App全名*", { type: 'text' }),
              this.renderInput("redirect_uri", "回调URL*", { type: 'url' }),
            ])
          }
          { this.renderAction("保存", this.onCreate) }
      </div>
    );
  },

  onCreate() {
    const fields = ['fullname', 'redirect_uri'];
    const {isValid, formData} = this.validateForm(fields, fields);
    if (isValid) {
      const {token, tokenType} = this.props;
      this.setState({ inRequest: true });
      Admin.createApplication(token, tokenType, formData, this.onRequestCallback);
    }
  },


});

export default AdminNewAppCard;
