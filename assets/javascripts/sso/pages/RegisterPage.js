import StyleSheet from 'react-style';
import React from 'react';

import UserRegisterCard from '../components/UserRegisterCard';

let RegisterPage = React.createClass({
  render() {
    return (
      <div className="mdl-grid">
        <div className="mdl-cell mdl-cell--6-col mdl-cell--8-col-tablet mdl-cell--4-col-phone">
          <UserRegisterCard />
        </div>
      </div>
    );
  }
});

export default RegisterPage;