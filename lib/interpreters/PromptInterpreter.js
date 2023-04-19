const colors = require("colors");
const Crypto = require("../Crypto");
const out = require("../Output");
const prompt = require("prompt-sync")();
const cliSelect = require("cli-select-2");
const Variable = require("../Variable");

function PromptInterpreter() {
  return {
    interpret: async function (command, action, args, parameters) {
      let result = action;

      const variables = Variable.list(result);
      for (const name of variables) {
        if (command.help && command.help.variables) {
          for (const commandVariable of command.help.variables) {
            if (commandVariable.name === name) {
              if (commandVariable.default === undefined) {
                let value;

                if (commandVariable.options) {
                  value = await openOptions(commandVariable);
                } else {
                  value = await openPrompt(commandVariable);
                }

                parameters[commandVariable.name] = value;
                result = Variable.replace(result, name, value);
              }
            }
          }
        }
      }

      return result;
    }
  };
}

async function openPrompt(commandVariable) {
  const options = {};
  if (commandVariable.hide) {
    options.echo = "*";
  }

  const inputValue = prompt(displayText(commandVariable), options);
  return commandVariable.hide ? Crypto.encrypt(inputValue) : inputValue;
}

async function openOptions(commandVariable) {
  out.println(displayText(commandVariable));
  const response = await cliSelect({
    values: commandVariable.options
  });
  return response.value;
}

function displayText(commandVariable) {
  let text = commandVariable.name.bold;
  if (commandVariable.text) {
    text += ` [${commandVariable.text}]`;
  }
  return `${text}: `.cyan;
}

module.exports = PromptInterpreter;
