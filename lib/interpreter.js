const colors = require('colors');
const prompt = require('prompt-sync')();

const params = require('./params');

const PARAM_REGEX = /\$\{?(\w+)\}?/g;
const VARIABLE_REGEX = /\$\{?(\w+)\}?/;

module.exports = {
  interpret: function(command, action, args, parameters) {
    let result = action;

    let variables = result.match(PARAM_REGEX);
    if (variables) {
      variables.forEach(function(variable) {
        let key = variable.match(VARIABLE_REGEX)[1];
        let value = parameters[key];

        if (!value) {
          if (command.help && command.help.variables) {
            command.help.variables.forEach(function(commandVariable) {
              if (commandVariable.name === key) {
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

        result = result.replace(variable, value);
      });
    }

    return result;
  }
};
