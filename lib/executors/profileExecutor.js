const executor = require('../executor');

const PREFIX = 'profile:';

module.exports = {
  execute: function(command, action, args, parameters) {
    if (!action.startsWith(PREFIX)) {
      return false;
    }

    const profile = action.substring(PREFIX.length);
    executor.defineProfile(profile);

    executor.execute(args, parameters);

    return true;
  }
};
