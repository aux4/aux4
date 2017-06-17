const Cryptr = require('cryptr');

const SECURITY_KEY = 'C1E111867141295954C8DF64426FD';
const AUX4_SECURITY_KEY = process.env.AUX4_SECURITY_KEY || SECURITY_KEY;
const cryptr = new Cryptr(AUX4_SECURITY_KEY);

module.exports = {
  encrypt: function(text) {
    return cryptr.encrypt(text);
  },
  decrypt: function(key) {
    return cryptr.decrypt(key);
  }
};
