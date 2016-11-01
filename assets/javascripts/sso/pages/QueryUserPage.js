import StyleSheet from 'react-style';
import React from 'react';

import UserPublicInfoCard from '../components/UserPublicInfoCard';
import QueryUserCard from '../components/QueryUserCard';

let QueryUserPage = React.createClass({
  render(){
    const {params}=this.props;
    console.log(params.name);
    return (
      <div className="mdl-grid">
        <div className="mdl-cell mdl-cell--6-col mdl-cell--8-col-tablet mdl-cell--4-col-phone">
          <QueryUserCard />
        </div>

        <div className="mdl-cell mdl-cell--12-col mdl-cell--8-col-tablet mdl-cell--4-col-phone">
          <UserPublicInfoCard user={params.name}/>
        </div>
      </div>
    );
  },

});

export default QueryUserPage;
