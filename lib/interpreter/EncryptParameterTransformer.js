const { Crypto } = require("@aux4/encrypt");

function encryptParameterTransformer(value, command, parameters, name) {
  const variable = command.help.variables.find(variable => variable.name === name);
  if (!variable || !variable.encrypt) {
    return value;
  }

  const crypto = new Crypto(parameters.secret);
  return crypto.encrypt(value);
}

module.exports = encryptParameterTransformer;
