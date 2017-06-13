const out = require('../output');

const interpreter = require('../interpreter');

const PREFIX = 'log:';

module.exports = {
  execute: function(command, action, args, parameters) {
    if (!action.startsWith(PREFIX)) {
      return false;
    }

    let text = action.substring(PREFIX.length);
    text = interpreter.interpret(command, text, args, parameters);
    out.println(text);

    return true;
  }
};
