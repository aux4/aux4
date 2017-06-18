const colors = require('colors');
const prompt = require('prompt-sync')();

const variable = require('../variable');

module.exports = {
  interpret: function(command, action, args, parameters) {
    let result = action;

    let variables = variable.list(result);
    variables.forEach(function(name) {
      let value = parameters[name];

      if (!value) {
        if (command.help && command.help.variables) {
          command.help.variables.forEach(function(commandVariable) {
            if (commandVariable.name === name) {
              if (commandVariable.default) {
                value = commandVariable.default;
              } else if (commandVariable.text) {
                value = prompt(commandVariable.text.cyan + ': ');
              }
            }
          });
        }
      }

      if (!value) {
        value = '';
      }

      result = variable.replace(result, name, value);
    });

    return result;
  }
};
