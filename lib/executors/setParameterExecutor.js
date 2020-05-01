const out = require('../output');
const interpreter = require('../interpreter');

const PREFIX = 'set:';

module.exports = {
  execute: function(command, action, args, parameters) {
    if (!action.startsWith(PREFIX)) {
      return false;
    }

    const parameterDeclaration = action.substring(PREFIX.length);

    const equalSignPosition = parameterDeclaration.indexOf('=');
    if (equalSignPosition === -1) {
      out.println('The set format is: set:<param-name>=<param-value>'.red);
      return true;
    }

    const parameterName = parameterDeclaration.substring(0, equalSignPosition);
    let parameterValue = parameterDeclaration.substring(equalSignPosition + 1);
    parameterValue = interpreter.interpret(command, parameterValue, args, parameters);

    parameters[parameterName] = parameterValue;

    return true;
  }
};
