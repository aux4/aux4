const { Crypto } = require("@aux4/encrypt");

const REGEX = /decrypt\(\S+\)/g;
const REGEX_CONTENT = /decrypt\((\S+)\)/;

class DecryptInterpreter {
  async interpret(command, action, args, parameters) {
    if (typeof action !== "string") {
      return action;
    }

    const secret = await parameters.secret;
    const crypto = new Crypto(secret);

    let result = action;

    const matches = result.match(REGEX);
    if (matches) {
      matches.forEach(function (match) {
        const hash = match.match(REGEX_CONTENT)[1];
        result = result.replace(match, crypto.decrypt(hash));
      });
    }
    return result;
  }
}

module.exports = DecryptInterpreter;
