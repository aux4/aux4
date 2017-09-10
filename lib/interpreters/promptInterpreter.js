const colors = require('colors');
const prompt = require('prompt-sync')();

const variable = require('../variable');

module.exports = {
  interpret: function(command, action, args, parameters) {
    let result = action;

    let variables = variable.list(result);
    variables.forEach(function(name) {
      if (command.help && command.help.variables) {
        command.help.variables.forEach(function(commandVariable) {
          if (commandVariable.name === name) {
            if (commandVariable.text && !commandVariable.default) {
              let value = prompt((commandVariable.text + ': ').cyan);
              parameters[commandVariable.name] = value;
              result = variable.replace(result, name, value);
            }
          }
        });
      }
    });

    return result;
  }
};
