let AdminAuthorizeMixin = {

  getInitialState() {
    const [token, tokenType] = this.getTokens();
    return {
      token,
      tokenType,
    };
  },

  authorize(area, ag) {
    if (!this.isSessionValid()) {
      let query = { area }
      if (ag) {
        query['ag'] = ag;
      }
      this.history.pushState(null, '/spa/admin/authorize', query);
    }
  },

  isSessionValid() {
    const {token, tokenType} = this.state;
    return token && tokenType;
  },

  getTokens() {
    const {state} = this.props.location;
    if (state) {
      const {token, tokenType} = state;
      return [token, tokenType];
    };
    return ['', ''];
  },

};

export default AdminAuthorizeMixin;
