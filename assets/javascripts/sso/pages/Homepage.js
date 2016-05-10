import StyleSheet from 'react-style';
import React from 'react';

import UserRegisterCard from '../components/UserRegisterCard';
import UserResetPasswordCard from '../components/UserResetPasswordCard';
import UserResetPasswordByEmailCard from '../components/UserResetPasswordByEmailCard.js';
import AdminIndexCard from '../components/AdminIndexCard';

let Homepage = React.createClass({

  render() {
    return (
      <div className="mdl-grid">
        <div className="mdl-cell mdl-cell--4-col mdl-cell--8-col-tablet mdl-cell--4-col-phone">
          <UserRegisterCard />
        </div>

        <div className="mdl-cell mdl-cell--4-col mdl-cell--8-col-tablet mdl-cell--4-col-phone">
          <UserResetPasswordCard />
          <UserResetPasswordByEmailCard />
        </div>

        <div className="mdl-cell mdl-cell--4-col mdl-cell--8-col-tablet mdl-cell--4-col-phone">
          <AdminIndexCard />
        </div>
      </div>
    );
  },

});

export default Homepage;
