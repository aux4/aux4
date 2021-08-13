const colors = require('colors');
const crypto = require('../crypto');
const prompt = require('prompt-sync')();

const variable = require('../variable');

module.exports = {
  interpret: function(command, action, args, parameters) {
    let result = action;

    const variables = variable.list(result);
    variables.forEach((name) => {
      if (command.help && command.help.variables) {
        command.help.variables.forEach((commandVariable) => {
          if (commandVariable.name === name) {
            if (commandVariable.default === undefined) {
              const options = {};
              if (commandVariable.hide) {
                options.echo = '*';
              }

              let text = commandVariable.name.bold;
              if (commandVariable.text) {
                text += ` [${commandVariable.text}]`;
              }

              const inputValue = prompt(`${text}: `.cyan, options);
              const value = commandVariable.hide ? crypto.encrypt(inputValue) : inputValue;

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
