const colors = require('colors');

const crypto = require('../crypto');
const out = require('../output');

const PREFIX = 'crypto:encrypt';

module.exports = {
  execute: function(command, action, args, parameters) {
    if (!action.startsWith(PREFIX)) {
      return false;
    }

    if (args.length === 0) {
      out.println('There is nothing to encrypt'.red);
      return true;
    }

    out.println(crypto.encrypt(args[0]));

    return true;
  }
};
