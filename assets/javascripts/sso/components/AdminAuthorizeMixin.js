import {Admin} from '../models/Models';
import Cookie from '../models/Cookie';
import {kCookieToken} from '../models/Models';

let AdminAuthorizeMixin = {

  getInitialState() {
    const [token, tokenType] = this.getTokens();
    return {
      token,
      tokenType,
    };
  },

  componentWillMount(){
    const {token, tokenType} = this.state;
    Admin.checkToken(token, tokenType, this.updateToken)
  },

  updateToken(newState){
    if(typeof newState.token === 'undefined'){
      Cookie.del(kCookieToken);
      this.setState({token:"", tokenType:""});
    }
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
    // since getInitialState is before componentWillMount in which we call authorize, state can be null unexpectly
    const {state} = this.props.location;
    if(state == null){
      const {access, ty} = Admin.getTokenCookie();
      if (access && ty) {
        const token = access;
        const tokenType = ty;
        return [token, tokenType];
      }
      else{
        return ['', ''];
      }
    }
    else {
      const {token, tokenType} = state;
      return [token, tokenType];
    }
  },

};

export default AdminAuthorizeMixin;
