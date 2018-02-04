const colors = require('colors');

const executorChain = require('./executorChain');
const Profile = require('./profile');
const help = require('./help');
const out = require('./output');
const defaultSuggester = require('./suggester');

let profiles;
let currentProfile = 'main';

let suggester = defaultSuggester;

module.exports = {
  init: function(config) {
    let cfg = config.get();
    let cfgProfiles = cfg.profiles;

    profiles = {};

    cfgProfiles.forEach(function(cfgProfile) {
      profiles[cfgProfile.name] = new Profile(config, cfgProfile.name);
    });
  },

  suggester: function(newSuggester) {
    suggester = newSuggester;
  },

  profile: function(name) {
    return profiles[name];
  },

  defineProfile: function(name) {
    if (profiles[name] === undefined) {
      throw new Error(`profile ${name} does not exists`);
    }
    currentProfile = name;
  },

  currentProfile: function() {
    return currentProfile;
  },

  execute: function(args, parameters = {}) {
    let profile = profiles[currentProfile];

    if (args.length === 0) {
      let length = maxLength(profile.commands());
      profile.commands().forEach(function(command) {
        help.print(command, length + 2);
      });
      return;
    }

    let command = profile.command(args[0]);
    if (!command) {
      suggester.suggest(profile, args[0]);
      return;
    }

    if (args.length === 1 && parameters.help) {
      help.print(command, command.value.length + 2);
      return;
    }

    executorChain.execute(command, args.splice(1), parameters);
  }
};

function maxLength(commands) {
  let maxLength = 0;

  commands.forEach(function(command) {
    if (command.value.length > maxLength) {
      maxLength = command.value.length;
    }
  });

  return maxLength;
}
