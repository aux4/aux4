const crypto = require('../crypto');

const REGEX = /crypto\(\S+\)/g;
const REGEX_CONTENT = /crypto\((\S+)\)/;

module.exports = {
  interpret: function(command, action, args, parameters) {
    let result = action;

    let matches = result.match(REGEX);
    if (matches) {
      matches.forEach(function(match) {
        let hash = match.match(REGEX_CONTENT)[1];
        result = result.replace(match, crypto.decrypt(hash));
      });
    }

    return result;
  }
};
