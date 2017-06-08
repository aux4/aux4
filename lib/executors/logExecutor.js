const out = require('../output');

const interpreter = require('../interpreter');

const PREFIX = 'log:';

module.exports = {
  execute: function(command, args) {
    if (!command.startsWith(PREFIX)) {
      return false;
    }

    let text = command.substring(PREFIX.length);
    text = interpreter.interpret(text, args);
    out.println(text);

    return true;
  }
};
