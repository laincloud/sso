import StyleSheet from 'react-style';
import React from 'react';

export function createElement(Component) {
  const SsoAppComponent = React.createClass({
    componentDidMount() {
      componentHandler.upgradeDom();
    },

    componentDidUpdate() {
      componentHandler.upgradeDom();
    }, 

    render() {
      return (
        <div style={this.styles.pageContent}>
          <Component {...this.props} />
        </div>
      ); 
    },

    styles: StyleSheet.create({
      pageContent: {
        display: "block",
        paddingTop: 16,
        maxWidth: 960,
        margin: "0 auto",
      },
    }),
  });

  return SsoAppComponent;
}

