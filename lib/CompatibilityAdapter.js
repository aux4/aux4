class CompatibilityAdapter {
  static adapt(config) {
    (config.profiles || []).forEach(profile => {
      (profile.commands || []).forEach(command => {
        if (command.name === undefined) {
          command.name = command.value;
        }

        if (command.help) {
          if (command.help.text === undefined) {
            command.help.text = command.help.description;
          }
        }
      });
    });
  }
}

module.exports = CompatibilityAdapter;
