const variable = require("../Variable");

function DefaultInterpreter() {
  return {
    interpret: function (command, action, args, parameters) {
      let result = action;

      const variables = variable.list(result);
      variables.forEach(function (name) {
        if (command.help && command.help.variables) {
          command.help.variables.forEach(function (commandVariable) {
            if (commandVariable.name === name) {
              if (commandVariable.default) {
                const value = commandVariable.default;
                result = variable.replace(result, name, value);
              }
            }
          });
        }
      });

      return result;
    }
  };
}

module.exports = DefaultInterpreter;
