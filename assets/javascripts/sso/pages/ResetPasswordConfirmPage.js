import StyleSheet from 'react-style';
import React from 'react';
import {History} from 'react-router';

import UserResetPasswordConfirmCard from '../components/UserResetPasswordConfirmCard';

let ResetPasswordConfirmPage = React.createClass({
  mixins: [History],

  render() {
    const {params} = this.props;
    return (
      <div className="mdl-grid">
        <div className="mdl-cell mdl-cell--6-col mdl-cell--8-col-tablet mdl-cell--4-col-phone">
          <UserResetPasswordConfirmCard code={params.code} username={params.username}
            onConfirm={this.onConfirm} />
        </div>
      </div>
    );
  },

  onConfirm() {
    this.history.replaceState(null, '/spa/', null);
  },

});

export default ResetPasswordConfirmPage;
