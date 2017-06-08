const executor = require('../executor');

const PREFIX = 'profile:';

module.exports = {
  execute: function(command, args) {
    if (!command.startsWith(PREFIX)) {
      return false;
    }

    let profile = command.substring(PREFIX.length);
    executor.defineProfile(profile);

    executor.execute(args);

    return true;
  }
};
