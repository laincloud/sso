import StyleSheet from 'react-style';
import React from 'react';

import {User} from '../models/Models';
import CardFormMixin from './CardFormMixin';
import {History} from 'react-router'

let QueryUserCard = React.createClass({
  mixins: [CardFormMixin, History],

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
          <h2 className="mdl-card__title-text">用户查询</h2>
        </div>
        { this.renderResult() }
        {
          reqResult.fin && reqResult.ok ? null :
            this.renderForm(this.onQuery, [
              this.renderInput("name", "用户名*(字母、数字和减号)", { type: "text", pattern: "[\-a-zA-Z0-9]*" }), 
            ])
        }
        { this.renderAction("查询", this.onQuery) }
      </div>
    );
  },

  onQuery() {
    const {isValid, formData} = this.validateForm(["name"], ["name"]);
    if (isValid) {
      name=formData['name'];
      //this.setState({ inRequest: true });
      this.history.pushState(null,`/spa/users-query/${name}`)
    }
  },

});

export default QueryUserCard;
