const colors = require("colors");
const Crypto = require("../Crypto");
const prompt = require("prompt-sync")();

const Variable = require("../Variable");

function PromptInterpreter() {
  return {
    interpret: function (command, action, args, parameters) {
      let result = action;

      const variables = Variable.list(result);
      variables.forEach(name => {
        if (command.help && command.help.variables) {
          command.help.variables.forEach(commandVariable => {
            if (commandVariable.name === name) {
              if (commandVariable.default === undefined) {
                const options = {};
                if (commandVariable.hide) {
                  options.echo = "*";
                }

                let text = commandVariable.name.bold;
                if (commandVariable.text) {
                  text += ` [${commandVariable.text}]`;
                }

                const inputValue = prompt(`${text}: `.cyan, options);
                const value = commandVariable.hide ? Crypto.encrypt(inputValue) : inputValue;

                parameters[commandVariable.name] = value;
                result = Variable.replace(result, name, value);
              }
            }
          });
        }
      });

      return result;
    }
  };
}

module.exports = PromptInterpreter;
