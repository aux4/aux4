const colors = require("colors");

const out = require("./Output");

const Help = {
  print: function (command, length) {
    if (!length) {
      length = command.value.length;
    }

    const commandName = leftPad(command.value, length, " ");
    const description = command.help ? command.help.description : "";
    out.println(commandName.yellow, " ", indentParagraphs(description, length + 3));

    if (command.help && command.help.variables) {
      command.help.variables.forEach(variable => {
        let defaultValue = "";
        if (variable.default) {
          defaultValue = `[${variable.default.italic}]`;
        }
        out.println(
          leftPad("-", length + 6, " "),
          variable.name.cyan,
          defaultValue,
          indentParagraphs(variable.text, length + 7)
        );
      });
    }
  }
};

function indentParagraphs(text = "", length) {
  return text.replace(/\n/, "\n" + leftPad("", length, " "));
}

function leftPad(text, length, char) {
  while (text.length < length) {
    text = char + text;
  }
  return text;
}

module.exports = Help;
