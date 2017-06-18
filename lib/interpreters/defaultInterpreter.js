const variable = require('../variable');

module.exports = {
  interpret: function(command, action, args, parameters) {
    let result = action;

    let variables = variable.list(result);
    variables.forEach(function(name) {
      if (command.help && command.help.variables) {
        command.help.variables.forEach(function(commandVariable) {
          if (commandVariable.name === name) {
            if (commandVariable.default) {
              let value = commandVariable.default;
              result = variable.replace(result, name, value);
            }
          }
        });
      }
    });

    return result;
  }
};
