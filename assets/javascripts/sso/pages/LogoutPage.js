import React from 'react';
import {History} from 'react-router';
import Cookie from '../models/Cookie';
import {kCookieToken} from '../models/Models';

let LogoutPage = React.createClass({

  mixins: [History],

  componentDidMount() {
    Cookie.del(kCookieToken); 
    this.history.pushState(null, '/spa');
  },

  render() {
    return (
      <div>退出中......</div>
    ); 
  },

});

export default LogoutPage;
