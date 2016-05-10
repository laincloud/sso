import StyleSheet from 'react-style';
import React from 'react';
import classnames from 'classnames';

let CardFormMixin = {

  getInitialState() {
    return {
      inRequest: false,
      reqResult: {
        fin: false,
        ok: false,
        status: '',
      },
    };
  },

  componentDidUpdate() {
    componentHandler.upgradeDom();
  },

  onRequestCallback(ok, status) {
    this.setState({
      inRequest: false,
      reqResult: { fin: true, ok, status },
    }, () => {
      if (ok) {
        this.props.onSucc && this.props.onSucc();
      }
    });
    setTimeout(() => {
      this.setState({
        reqResult: { fin: false, ok: false, status: '' },
      });
    }, 5000);
  },

  validateForm(allFields, requiredFields) {
    const fields = allFields;
    let formData = {};
    _.forEach(fields, (field) => {
      let node = this.refs[field].getDOMNode();
      if (node.value && node.checkValidity()) {
        formData[field] = node.value;
      }
    });
    let formValids = _.zipObject(_.map(requiredFields, (field) => {
      return [field, Boolean(formData[field])];
    }));

    let isValid = true;
    if (!_.all(formValids)) {
      this.setState({ formValids });
      isValid = false;
    }
    return {isValid, formData};
  },

  renderAction(title, callback) {
    const {inRequest, reqResult} = this.state;
    if (reqResult.fin && reqResult.ok) {
      return null;
    }
    let child = <div className="mdl-spinner mdl-js-spinner is-active" style={{ marginRight: 24 }}></div>;
    if (!inRequest) {
      child = (
        <button className="mdl-button mdl-js-button mdl-js-ripple-effect mdl-button--colored" 
          disabled={this.state.inRequest}
          onClick={callback}>{title}</button>
      );
    }
    return (
      <div className="mdl-card__actions" style={{ textAlign: 'right' }}>
        { child }    
      </div>
    );
  },

  renderForm(submitCallback, inputs) {
    const checkKeyPress = (evt) => {
      if (evt && evt.keyCode === 13) {
        submitCallback && submitCallback();
        evt.preventDefault();
      } 
    }
    return (
      <form style={this.styles.form} onKeyDown={checkKeyPress}>
        { inputs }
      </form>
    );
  },

  renderInput(name, title, props) {
    const id = `txt-${name}`;
    const {formValids} = this.state;
    let extraClazz = {
      'is-invalid': !formValids[name],
      'is-dirty': !formValids[name],
    };
    return (
      <div className={classnames("mdl-textfield mdl-js-textfield mdl-textfield--floating-label", extraClazz)} 
          style={{ width: '100%' }}>
        <input className="mdl-textfield__input" id={id} ref={name} {...props} />
        <label className="mdl-textfield__label" htmlFor={id}>{title}</label>
      </div>
    );
  },

  renderResult() {
    const {reqResult} = this.state;
    if (!reqResult.fin) {
      return null;
    }
    return (
      <div className="mdl-card__supporting-text" style={{ 
          color: reqResult.ok ? "#4CAF50" : "#F44336",
          borderTop: '1px solid rgba(0, 0, 0, .1)',
        }}>
        {reqResult.status}
      </div>
    );
  },

  styles: StyleSheet.create({
    card: {
      width: '100%',
      marginBottom: 16,
      minHeight: 50,
    },
    form: {
      padding: '0 16px',
    },
  }),
}

export default CardFormMixin;
