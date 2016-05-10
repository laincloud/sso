import cookie from 'cookie';

const Cookie = {
  
  get(key) {
    const cookies = cookie.parse(document.cookie);
    return cookies ? cookies[key] : undefined;
  },

  set(key, value, options) {
    const objValue = typeof value === 'object' ? 
      JSON.stringify(value) : value;
    document.cookie = cookie.serialize(key, objValue, options);
  },

  del(key, path) {
    let options = {
      path: path || '/',
      expires: new Date(0),
    };
    this.set(key, "", options);
  },

  has(key) {
    return this.get(key) !== undefined; 
  },

  keys() {
    const cookies = cookie.parse(document.cookie); 
    let keys = [];
    for (let c in cookies) {
      keys.push(c);
    }
    return keys;
  },

  runTests() {
    const key = "test_haha";
    this.set(key, "hello world", {
      path: '/',
      maxAge: 50000,
    });
    const value = this.get(key);
    if (value !== "hello world") {
      console.error("Error, should find the cookie");
    }
    if (!this.has(key)) {
      console.error("Error, we should have the cookie");
    }

    this.del(key);
    const keys = this.keys();
    for (let k in keys) {
      console.log(keys[k]);
      if (keys[k] === key) {
        console.error("We should not find this cookie after delete");
      }
    }
  },

};

export default Cookie;
