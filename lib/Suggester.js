const out = require("./Output");

const Suggester = {
  suggest: function (profile, value) {
    const suggestions = [];

    profile.commands().forEach(function (cmd) {
      if (cmd.value.startsWith(value)) {
        suggestions.push(cmd.value);
      }
    });

    if (suggestions.length === 0) {
      out.println(`Command not found: ${value}`);
    } else {
      out.println("What did you mean:");
      suggestions.forEach(function (suggestion) {
        out.println("  - ", suggestion.bold);
      });
    }
  }
};

module.exports = Suggester;
