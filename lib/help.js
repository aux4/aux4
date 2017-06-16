const colors = require('colors');

const out = require('./output');

module.exports = {
  print: function(command, length) {
    if (!length) {
      length = command.value.length;
    }

    let commandName = leftPad(command.value, length, ' ');
    let description = command.help ? command.help.description : '';
    out.println(commandName.yellow, ' ', description);

    if (command.help && command.help.variables) {
      command.help.variables.forEach(function(variable) {
        let defaultValue = '';
        if (variable.default) {
          defaultValue = `[${variable.default.italic}]`;
        }
        out.println(leftPad('-', length + 6, ' '), variable.name.bold, defaultValue, variable.text);
      });
    }
  }
};

function leftPad(text, length, char) {
  while (text.length < length) {
    text = char + text;
  }
  return text;
}
