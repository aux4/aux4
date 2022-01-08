const out = require("../Output");

const PREFIX = "log:";

function LogExecutor(interpreter) {
  return {
    execute: function (command, action, args, parameters) {
      if (!action.startsWith(PREFIX)) {
        return false;
      }

      let text = action.substring(PREFIX.length);
      text = interpreter.interpret(command, text, args, parameters);
      out.println(text);

      return true;
    }
  };
}

module.exports = LogExecutor;
