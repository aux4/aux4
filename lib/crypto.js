const Cryptr = require('cryptr');

const SECURITY_KEY = 'C1E1118671412959';
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
