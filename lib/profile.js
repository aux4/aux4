const config = require('./config');

module.exports = function(name) {
  let profile;

  config.file().profiles.forEach(function(item){
    if (name === item.name) {
      profile = item;
    }
  });

  if (profile === undefined) {
    throw new Error(`profile ${name} not found in the configuration file`);
  }

  return {
    name: function() {
      return name;
    },

    commands: function() {
      return profile.commands;
    },

    command: function(name) {
      let selected;
      profile.commands.forEach(function(command) {
        if (name === command.value) {
          selected = command;
        }
      });
      return selected;
    }
  };
}
