import 'es5-shim';
import 'whatwg-fetch';
import _ from 'lodash';
window._ = _;

import StyleSheet from 'react-style';
import React from 'react';
import {Router, Route} from 'react-router';
import createBrowserHistory from 'history/lib/createBrowserHistory'

import {createElement} from './sso/app';
import Homepage from './sso/pages/Homepage';
import RegisterPage from './sso/pages/RegisterPage';
import LogoutPage from './sso/pages/LogoutPage';
import ResetPasswordPage from './sso/pages/ResetPasswordPage';
import ResetPasswordConfirmPage from './sso/pages/ResetPasswordConfirmPage';
import AdminAuthorizePage from './sso/pages/AdminAuthorizePage';
import AdminAppsPage from './sso/pages/AdminAppsPage';
import AdminGroupsPage from './sso/pages/AdminGroupsPage';
import AdminMembersPage from './sso/pages/AdminMembersPage';
import AdminUsersPage from './sso/pages/AdminUsersPage';

let domReady = () => {
  React.initializeTouchEvents(true);

  let history = createBrowserHistory();
  React.render((
    <Router history={history} >
      <Route path="/spa" component={createElement(Homepage)} />

      <Route path="/spa/user/register" component={createElement(RegisterPage)} />
      <Route path="/spa/user/password/reset" component={createElement(ResetPasswordPage)} />
      <Route path="/spa/user/password/reset/:username/:code" component={createElement(ResetPasswordConfirmPage)} />
      <Route path="/spa/user/logout" component={createElement(LogoutPage)} />

      <Route path="/spa/admin/authorize" component={createElement(AdminAuthorizePage)} />
      <Route path="/spa/admin/apps" component={createElement(AdminAppsPage)} />
      <Route path="/spa/admin/groups" component={createElement(AdminGroupsPage)} />
      <Route path="/spa/admin/groups/:name" component={createElement(AdminMembersPage)} />
      <Route path="/spa/admin/users" component={createElement(AdminUsersPage)} />
    </Router>
  ), document.getElementById("sso-spa"));
};

if (typeof document.onreadystatechange === "undefined") {
    window.onload = () => domReady();
} else {
    document.onreadystatechange = () => {
      if (document.readyState !== "complete") {
        return;
      }
      domReady();
    }
}
